package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
	log "github.com/shylinux/toolkits/logs"

	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const LOGIN = "_login"

func _serve_login(msg *ice.Message, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	msg.Option(ice.MSG_USERNAME, "")
	msg.Option(ice.MSG_USERROLE, "void")

	if msg.Options(ice.MSG_SESSID) {
		// 会话认证
		aaa.SessCheck(msg, msg.Option(ice.MSG_SESSID))
	}

	if !msg.Options(ice.MSG_USERNAME) && tcp.IPIsLocal(msg, msg.Option(ice.MSG_USERIP)) {
		// 自动认证
		if aaa.UserLogin(msg, cli.UserName, cli.PassWord) {
			if strings.HasPrefix(msg.Option(ice.MSG_USERUA), "Mozilla/5.0") {
				// msg.Option(ice.MSG_SESSID, aaa.SessCreate(msg, msg.Option(ice.MSG_USERNAME), msg.Option(ice.MSG_USERROLE)))
				// Render(msg, "cookie", msg.Option(ice.MSG_SESSID))
			}
		}
	}

	if _, ok := msg.Target().Commands[LOGIN]; ok {
		// 权限检查
		msg.Target().Cmd(msg, LOGIN, msg.Option(ice.MSG_USERURL), cmds...)
		cmds = kit.Simple(msg.Optionv(ice.MSG_CMDS))

	} else if ls := strings.Split(msg.Option(ice.MSG_USERURL), "/"); msg.Conf(SERVE, kit.Keys("meta.black", ls[1])) == "true" {
		return cmds, false // black

	} else if msg.Conf(SERVE, kit.Keys("meta.white", ls[1])) == "true" {
		return cmds, true // white

	} else {
		if msg.Warn(!msg.Options(ice.MSG_USERNAME), "not login %s", msg.Option(ice.MSG_USERURL)) {
			msg.Render(STATUS, 401, "not login")
			return cmds, false
		}
		if msg.Warn(!msg.Right(msg.Option(ice.MSG_USERURL))) {
			msg.Render(STATUS, 403, "not auth")
			return cmds, false
		}
	}

	return cmds, msg.Option(ice.MSG_USERURL) != ""
}
func _serve_handle(key string, cmd *ice.Command, msg *ice.Message, w http.ResponseWriter, r *http.Request) {
	defer func() { msg.Cost("%s %v %v", r.URL.Path, msg.Optionv(ice.MSG_CMDS), msg.Format("append")) }()

	// 请求变量
	msg.Option(ice.MSG_SESSID, "")
	for _, v := range r.Cookies() {
		msg.Option(v.Name, v.Value)
	}

	// 请求
	msg.Option(ice.MSG_OUTPUT, "")
	if u, e := url.Parse(r.Header.Get("Referer")); e == nil {
		for k, v := range u.Query() {
			if msg.Logs("refer", k, v); k != "name" {
				msg.Option(k, v)
			}
		}
	}

	// 用户请求
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
	switch r.Header.Get("Content-Type") {
	case "application/json":
		var data interface{}
		if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e != nil, "%s", e) {
			msg.Optionv(ice.MSG_USERDATA, data)
			msg.Logs("json", "value", kit.Formats(data))
		}

		switch d := data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				msg.Optionv(k, v)
			}
		}
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select(r.Header.Get("Content-Length"), "4096")))
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
func _serve_main(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("index.module") == "" {
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

		r.Header.Set("index.module", m.Target().Name)
		r.Header.Set("index.path", r.URL.Path)
		r.Header.Set("index.url", r.URL.String())

		if m.Conf(SERVE, "meta.logheaders") == "true" {
			// 请求参数
			for k, v := range r.Header {
				m.Info("%s: %v", k, kit.Format(v))
			}
			m.Info(" ")

			defer func() {
				// 响应参数
				for k, v := range w.Header() {
					m.Info("%s: %v", k, kit.Format(v))
				}
				m.Info(" ")
			}()
		}
	}

	if strings.HasPrefix(r.URL.Path, "/debug") {
		r.URL.Path = strings.Replace(r.URL.Path, "/debug", "/code", -1)
	}

	if r.URL.Path == "/" && m.Conf(SERVE, "meta.init") != "true" && len(ice.BinPack) == 0 {
		if _, e := os.Stat(m.Conf(SERVE, "meta.volcanos.path")); e == nil {
			// 初始化成功
			m.Conf(SERVE, "meta.init", "true")
		}
		m.W = w
		Render(m, "refresh", m.Conf(SERVE, "meta.volcanos.refresh"))
		m.Event(gdb.SYSTEM_INIT)
		m.W = nil
	} else if r.URL.Path == "/" && m.Conf(SERVE, "meta.sso") != "" {
		if r.ParseForm(); r.FormValue(ice.MSG_SESSID) != "" {
			return true
		}

		sessid := ""
		if c, e := r.Cookie(ice.MSG_SESSID); e == nil {
			m.Richs("aaa.sess", "", c.Value, func(key string, value map[string]interface{}) {
				sessid = c.Value
			})
		}

		if sessid == "" {
			http.Redirect(w, r, m.Conf(SERVE, "meta.sso"), http.StatusTemporaryRedirect)
			return false
		}
		return true
	} else if r.URL.Path == "/share" && r.Method == "GET" {
		http.ServeFile(w, r, m.Conf(SERVE, "meta.page.share"))
	} else {
		if b, ok := ice.BinPack[r.URL.Path]; ok {
			log.Info("BinPack %v %v", r.URL.Path, len(b))
			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			}
			w.Write(b)
			return false
		}
		return true
	}
	return false
}

const SERVE = "serve"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVE: {Name: "serve", Help: "服务器", Value: kit.Data(
				"init", "false", "logheaders", "false",
				"black", kit.Dict(),
				"white", kit.Dict(
					"login", true,
					"share", true,
					"space", true,
					"route", true,
					"static", true,
					"plugin", true,
					"publish", true,
				),

				"static", kit.Dict("/", "usr/volcanos/"),
				"volcanos", kit.Dict("path", "usr/volcanos", "branch", "master",
					"repos", "https://github.com/shylinux/volcanos",
					"require", ".ish/pluged",
					"refresh", "5",
				), "page", kit.Dict(
					"index", "usr/volcanos/page/index.html",
					"share", "usr/volcanos/page/share.html",
				), "publish", "usr/publish/",

				"template", kit.Dict("path", "usr/template", "list", []interface{}{
					`{{define "raw"}}{{.Result}}{{end}}`,
				}),
			)},
		},
		Commands: map[string]*ice.Command{
			SERVE: {Name: "serve [random] [ups...]", Help: "服务器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if cli.NodeInfo(m, SERVER, cli.HostName); len(arg) > 0 && arg[0] == "random" {
					cli.NodeInfo(m, SERVER, cli.PathName)
					// 随机端口
					m.Cmd(SPIDE, mdb.CREATE, "self", "http://random")
					arg = arg[1:]
				}

				// 启动服务
				m.Target().Start(m, "self")
				defer m.Cmd(SPACE, "connect", "self")
				m.Sleep("1s")

				// 连接服务
				for _, k := range arg {
					m.Cmd(SPACE, "connect", k)
				}
			}},
		}}, nil)
}
