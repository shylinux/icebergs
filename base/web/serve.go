package web

import (
	"io"
	"os"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func _serve_proxy(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	m.Option(SPIDE_CB, func(msg *ice.Message, req *http.Request, res *http.Response) {
		p := path.Join("var/proxy", strings.ReplaceAll(r.URL.String(), "/", "_"))
		size := 0
		if s, e := os.Stat(p); os.IsNotExist(e) {
			if f, p, e := kit.Create(p); m.Assert(e) {
				defer f.Close()

				if n, e := io.Copy(f, res.Body); m.Assert(e) {
					m.Debug("proxy %s res: %v", p, n)
					size = int(n)
				}
			}
		} else {
			size = int(s.Size())
		}

		h := w.Header()
		for k, v := range res.Header {
			for _, v := range v {
				switch k {
				case ContentLength:
					h.Add(k, kit.Format(size))
					m.Debug("proxy res: %v %v", k, size+1)
				default:
					m.Debug("proxy res: %v %v", k, v)
					h.Add(k, v)
				}
			}
		}
		w.WriteHeader(res.StatusCode)

		if f, e := os.Open(p); m.Assert(e) {
			defer f.Close()

			if n, e := io.Copy(w, f); e == nil {
				m.Debug("proxy res: %v", n)
			} else {
				m.Debug("proxy res: %v %v", n, e)
			}
		}
	})
	m.Cmdx(SPIDE, r.URL.Host, SPIDE_PROXY, r.Method, r.URL.String())
	return true
}
func _serve_main(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("index.module") != "" {
		return true
	}

	// 解析地址
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		r.Header.Set(ice.MSG_USERIP, ip)
	} else if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		r.Header.Set(ice.MSG_USERIP, ip)
	} else if strings.HasPrefix(r.RemoteAddr, "[") {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, "]")[0][1:])
	} else {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, ":")[0])
	}
	m.Info("").Info("%s %s %s", r.Header.Get(ice.MSG_USERIP), r.Method, r.URL)

	if strings.HasPrefix(r.URL.String(), "http") {
		if m == nil {
			m = ice.Pulse.Spawn()
		}
		return _serve_proxy(m, w, r)
	}

	// 请求地址
	r.Header.Set("index.module", m.Target().Name)
	r.Header.Set("index.path", r.URL.Path)
	r.Header.Set("index.url", r.URL.String())

	// 输出日志
	if m.Conf(SERVE, "meta.logheaders") == "true" {
		for k, v := range r.Header {
			m.Info("%s: %v", k, kit.Format(v))
		}
		m.Info(" ")

		defer func() {
			for k, v := range w.Header() {
				m.Info("%s: %v", k, kit.Format(v))
			}
			m.Info(" ")
		}()
	}

	if r.URL.Path == "/" && r.FormValue(SHARE) != "" {
		m.W = w
		defer func() { m.W = nil }()

		if s := m.Cmd(SHARE, mdb.SELECT, kit.MDB_HASH, r.FormValue(SHARE)); s.Append(kit.MDB_TYPE) == "login" {
			defer func() { http.Redirect(w, r, kit.MergeURL(r.URL.String(), SHARE, ""), http.StatusTemporaryRedirect) }()

			msg := m.Spawn()
			if c, e := r.Cookie(ice.MSG_SESSID); e == nil && c.Value != "" {
				if aaa.SessCheck(msg, c.Value); msg.Option(ice.MSG_USERNAME) != "" {
					return false // 复用会话
				}
			}

			msg.Option(ice.MSG_USERUA, r.Header.Get("User-Agent"))
			msg.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
			Render(msg, COOKIE, aaa.SessCreate(msg, s.Append(aaa.USERNAME), s.Append(aaa.USERROLE)))
			return false // 新建会话
		}

		return true
	}

	if strings.HasPrefix(r.URL.Path, "/debug") {
		r.URL.Path = strings.Replace(r.URL.Path, "/debug", "/code", -1)
		return true
	}

	if b, ok := ice.BinPack[r.URL.Path]; ok {
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		w.Write(b)
		return false
	}

	if r.URL.Path == "/" && strings.Contains(r.Header.Get("User-Agent"), "curl") {
		http.ServeFile(w, r, path.Join(m.Conf(SERVE, "meta.intshell.path"), m.Conf(SERVE, "meta.intshell.index")))
		return false
	}

	// 单点登录
	if r.URL.Path == "/" && m.Conf(SERVE, "meta.sso") != "" {
		sessid := r.FormValue(ice.MSG_SESSID)
		if sessid == "" {
			if c, e := r.Cookie(ice.MSG_SESSID); e == nil {
				sessid = c.Value
			}
		}

		ok := false
		m.Richs(aaa.SESS, "", sessid, func(key string, value map[string]interface{}) {
			ok = true
		})

		if !ok {
			http.Redirect(w, r, m.Conf(SERVE, "meta.sso"), http.StatusTemporaryRedirect)
			return false
		}
	}

	return true
}
func _serve_handle(key string, cmd *ice.Command, msg *ice.Message, w http.ResponseWriter, r *http.Request) {
	defer func() {
		msg.Cost(kit.Format("%s %v %v", r.URL.Path, msg.Optionv(ice.MSG_CMDS), msg.Format(ice.MSG_APPEND)))
	}()

	// 请求变量
	msg.Option(ice.MSG_SESSID, "")
	for _, v := range r.Cookies() {
		msg.Option(v.Name, v.Value)
	}

	// 请求地址
	if u, e := url.Parse(r.Header.Get("Referer")); e == nil {
		for k, v := range u.Query() {
			msg.Logs("refer", k, v)
			msg.Option(k, v)
		}
	}

	// 用户请求
	msg.Option(mdb.CACHE_LIMIT, "10")
	msg.Option(ice.MSG_OUTPUT, "")
	msg.Option(ice.MSG_METHOD, r.Method)
	msg.Option(ice.MSG_USERWEB, kit.Select(msg.Conf(SHARE, "meta.domain"), r.Header.Get("Referer")))
	msg.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	msg.Option(ice.MSG_USERUA, r.Header.Get("User-Agent"))
	msg.Option(ice.MSG_USERURL, r.URL.Path)
	if msg.R, msg.W = r, w; r.Header.Get("X-Real-Port") != "" {
		msg.Option(ice.MSG_USERADDR, msg.Option(ice.MSG_USERIP)+":"+r.Header.Get("X-Real-Port"))
	} else {
		msg.Option(ice.MSG_USERADDR, r.RemoteAddr)
	}

	// 解析引擎
	switch r.Header.Get(ContentType) {
	case ContentJSON:
		var data interface{}
		if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e != nil, e) {
			msg.Optionv(ice.MSG_USERDATA, data)
			msg.Logs("json", "value", kit.Format(data))
		}

		switch d := data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				msg.Optionv(k, v)
			}
		}
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select(r.Header.Get(ContentLength), "4096")))
		if r.ParseForm(); len(r.PostForm) > 0 {
			for k, v := range r.PostForm {
				msg.Logs("form", k, v)
			}
		}
	}

	// 请求参数
	for k, v := range r.Form {
		for i, p := range v {
			v[i], _ = url.QueryUnescape(p)
		}
		if msg.Optionv(k, v); k == ice.MSG_SESSID {
			msg.Render(COOKIE, v[0])
		}
	}

	// 请求命令
	if msg.Option(ice.MSG_USERPOD, msg.Option("pod")); msg.Optionv(ice.MSG_CMDS) == nil {
		if p := strings.TrimPrefix(msg.Option(ice.MSG_USERURL), key); p != "" {
			msg.Optionv(ice.MSG_CMDS, strings.Split(p, "/"))
		}
	}

	// 执行命令
	if cmds, ok := _serve_login(msg, kit.Simple(msg.Optionv(ice.MSG_CMDS)), w, r); ok {
		msg.Option("_option", msg.Optionv(ice.MSG_OPTION))
		msg.Target().Cmd(msg, key, msg.Option(ice.MSG_USERURL), cmds...)
	}

	// 渲染引擎
	_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
	Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
}
func _serve_login(msg *ice.Message, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	msg.Option(ice.MSG_USERROLE, aaa.VOID)
	msg.Option(ice.MSG_USERNAME, "")

	if msg.Options(ice.MSG_SESSID) {
		// 会话认证
		aaa.SessCheck(msg, msg.Option(ice.MSG_SESSID))
	}

	if !msg.Options(ice.MSG_USERNAME) && tcp.IsLocalHost(msg, msg.Option(ice.MSG_USERIP)) && msg.Conf(SERVE, "meta.localhost") == "true" {
		// 自动认证
		aaa.UserLogin(msg, ice.Info.UserName, ice.Info.PassWord)
	}

	if _, ok := msg.Target().Commands[LOGIN]; ok {
		// 权限检查
		msg.Target().Cmd(msg, LOGIN, msg.Option(ice.MSG_USERURL), cmds...)
		cmds = kit.Simple(msg.Optionv(ice.MSG_CMDS))

	} else if ls := strings.Split(msg.Option(ice.MSG_USERURL), "/"); msg.Conf(SERVE, kit.Keys("meta.black", ls[1])) == "true" {
		return cmds, false // 白名单

	} else if msg.Conf(SERVE, kit.Keys("meta.white", ls[1])) == "true" {
		return cmds, true // 黑名单

	} else {
		if msg.Warn(!msg.Options(ice.MSG_USERNAME), ice.ErrNotLogin, msg.Option(ice.MSG_USERURL)) {
			msg.Render(STATUS, 401, ice.ErrNotLogin)
			return cmds, false // 未登录
		}
		if msg.Warn(!msg.Right(msg.Option(ice.MSG_USERURL))) {
			msg.Render(STATUS, 403, ice.ErrNotAuth)
			return cmds, false // 未授权
		}
	}

	return cmds, msg.Option(ice.MSG_USERURL) != ""
}

const (
	LOGIN = "_login"
)
const (
	SERVE_START = "serve.start"
	SERVE_CLOSE = "serve.close"
)
const SERVE = "serve"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVE: {Name: SERVE, Help: "服务器", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME,
				"logheaders", "false",
				"localhost", "true",
				"black", kit.Dict(), "white", kit.Dict(
					"login", true, "space", true, "share", true, "plugin", true, "publish", true, "intshell", true,
				),

				"static", kit.Dict("/", "usr/volcanos/"),

				"volcanos", kit.Dict("refresh", "5",
					"share", "usr/volcanos/page/share.html",
					"path", "usr/volcanos", "require", ".ish/pluged",
					"repos", "https://github.com/shylinux/volcanos", "branch", "master",
				), "publish", "usr/publish/",

				"intshell", kit.Dict(
					"index", "index.sh",
					"path", "usr/intshell", "require", ".ish/pluged",
					"repos", "https://github.com/shylinux/volcanos", "branch", "master",
				),
			)},
		},
		Commands: map[string]*ice.Command{
			SERVE: {Name: "serve name auto start", Help: "服务器", Action: map[string]*ice.Action{
				gdb.START: {Name: "start dev= name=self proto=http host= port=9020", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if cli.NodeInfo(m, SERVER, ice.Info.HostName); m.Option(tcp.PORT) == "random" {
						m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
					}

					m.Target().Start(m, kit.MDB_NAME, m.Option(kit.MDB_NAME), tcp.HOST, m.Option(tcp.HOST), tcp.PORT, m.Option(tcp.PORT))
					m.Sleep("1s")

					for _, k := range kit.Split(m.Option(SPIDE_DEV)) {
						m.Cmd(SPACE, "connect", k)
					}
				}},
				aaa.WHITE: {Name: "white", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range arg {
						m.Conf(SERVE, kit.Keys(kit.MDB_META, aaa.WHITE, k), true)
					}
				}},
				aaa.BLACK: {Name: "black", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range arg {
						m.Conf(SERVE, kit.Keys(kit.MDB_META, aaa.BLACK, k), true)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,status,name,port,dev", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, SERVE, "", mdb.HASH, kit.MDB_NAME, arg)
			}},

			"/intshell/": {Name: "/intshell/", Help: "脚本", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Render(ice.RENDER_DOWNLOAD, path.Join(m.Conf(SERVE, "meta.intshell.path"), path.Join(arg...)))
			}},
		}})
}
