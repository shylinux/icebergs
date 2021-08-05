package web

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
)

func _serve_main(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("index.module") == "" {
		r.Header.Set("index.module", m.Prefix())
	} else { // 模块接口
		return true
	}

	// 用户地址
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

	// 参数日志
	if m.Conf(SERVE, kit.Keym(LOGHEADERS)) == ice.TRUE {
		for k, v := range r.Header {
			m.Info("%s: %v", k, kit.Format(v))
		}
		m.Info("")

		defer func() {
			for k, v := range w.Header() {
				m.Info("%s: %v", k, kit.Format(v))
			}
			m.Info("")
		}()
	}

	// 调试接口
	if strings.HasPrefix(r.URL.Path, "/debug") {
		r.URL.Path = strings.Replace(r.URL.Path, "/debug", "/code", -1)
		return true
	}

	// 主页接口
	if r.Method == SPIDE_GET && r.URL.Path == "/" {
		msg := m.Spawn()
		msg.W, msg.R = w, r
		repos := kit.Select(ice.INTSHELL, ice.VOLCANOS, strings.Contains(r.Header.Get("User-Agent"), "Mozilla/5.0"))
		Render(msg, ice.RENDER_DOWNLOAD, path.Join(m.Conf(SERVE, kit.Keym(repos, kit.SSH_PATH)), m.Conf(SERVE, kit.Keym(repos, kit.SSH_INDEX))))
		return false
	}

	// 文件接口
	if ice.Dump(w, r.URL.Path, func(name string) { RenderType(w, name, "") }) {
		return false
	}
	return true
}
func _serve_params(msg *ice.Message, path string) {
	switch ls := strings.Split(path, "/"); kit.Select("", ls, 1) {
	case "share":
		msg.Logs("refer", ls[1], ls[2])
		msg.Option(ls[1], ls[2])
	case "chat":
		switch kit.Select("", ls, 2) {
		case "pod":
			msg.Logs("refer", ls[2], ls[3])
			msg.Option(ls[2], ls[3])
		}
	}
}
func _serve_handle(key string, cmd *ice.Command, msg *ice.Message, w http.ResponseWriter, r *http.Request) {
	// 环境变量
	msg.Option(mdb.CACHE_LIMIT, "10")
	msg.Option(ice.MSG_OUTPUT, "")
	msg.Option(ice.MSG_SESSID, "")
	for _, v := range r.Cookies() {
		msg.Option(v.Name, v.Value)
	}

	// 请求变量
	_serve_params(msg, r.URL.Path)
	if u, e := url.Parse(r.Header.Get("Referer")); e == nil {
		_serve_params(msg, u.Path)
		for k, v := range u.Query() {
			msg.Logs("refer", k, v)
			msg.Option(k, v)
		}
	}

	// 请求地址
	msg.Option(ice.MSG_USERWEB, kit.Select(msg.Conf(SHARE, kit.Keym(kit.MDB_DOMAIN)), r.Header.Get("Referer")))
	msg.Option(ice.MSG_USERUA, r.Header.Get("User-Agent"))
	msg.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	if msg.R, msg.W = r, w; r.Header.Get("X-Real-Port") != "" {
		msg.Option(ice.MSG_USERADDR, msg.Option(ice.MSG_USERIP)+":"+r.Header.Get("X-Real-Port"))
	} else {
		msg.Option(ice.MSG_USERADDR, msg.Option(ice.MSG_USERIP))
	}

	// 请求数据
	switch r.Header.Get(ContentType) {
	case ContentJSON:
		var data interface{}
		if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e != nil, e) {
			msg.Log_IMPORT(kit.MDB_VALUE, kit.Format(data))
			msg.Optionv(ice.MSG_USERDATA, data)
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
		if r.Header.Get(ContentType) != ContentJSON {
			if msg.IsCliUA() {
				for i, p := range v {
					v[i], _ = url.QueryUnescape(p)
				}
			}
		}
		if msg.Optionv(k, v); k == ice.MSG_SESSID {
			RenderCookie(msg, v[0])
		}
	}

	// 请求命令
	if msg.Option(ice.MSG_USERPOD, msg.Option(cli.POD)); msg.Optionv(ice.MSG_CMDS) == nil {
		if p := strings.TrimPrefix(r.URL.Path, key); p != "" {
			msg.Optionv(ice.MSG_CMDS, strings.Split(p, "/"))
		}
	}

	// 执行命令
	if cmds, ok := _serve_login(msg, kit.Simple(msg.Optionv(ice.MSG_CMDS)), w, r); ok {
		msg.Option(ice.MSG_OPTS, msg.Optionv(ice.MSG_OPTION))
		msg.Target().Cmd(msg, key, r.URL.Path, cmds...)
		msg.Cost(kit.Format("%s %v %v", r.URL.Path, cmds, msg.Format(ice.MSG_APPEND)))
	}

	// 输出响应
	_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
	Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
}
func _serve_login(msg *ice.Message, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	msg.Option(ice.MSG_USERROLE, aaa.VOID)
	msg.Option(ice.MSG_USERNAME, "")

	if msg.Option(ice.MSG_SESSID) != "" {
		aaa.SessCheck(msg, msg.Option(ice.MSG_SESSID))
		// 会话认证
	}

	if msg.Option(ice.MSG_USERNAME) == "" && tcp.IsLocalHost(msg, msg.Option(ice.MSG_USERIP)) && msg.Conf(SERVE, kit.Keym(tcp.LOCALHOST)) == ice.TRUE {
		aaa.UserRoot(msg)
		// 主机认证
	}

	if _, ok := msg.Target().Commands[WEB_LOGIN]; ok {
		// 权限检查
		msg.Target().Cmd(msg, WEB_LOGIN, r.URL.Path, cmds...)
		return cmds, msg.Result(0) != ice.ErrWarn && msg.Result() != ice.FALSE
	}

	if ls := strings.Split(r.URL.Path, "/"); msg.Conf(SERVE, kit.Keym(aaa.BLACK, ls[1])) == ice.TRUE {
		return cmds, false // 黑名单
	} else if msg.Conf(SERVE, kit.Keym(aaa.WHITE, ls[1])) == ice.TRUE {
		if msg.Option(ice.MSG_USERNAME) == "" && msg.Option(SHARE) != "" {
			share := msg.Cmd(SHARE, msg.Option(SHARE))
			switch share.Append(kit.MDB_TYPE) {
			case LOGIN:
				// Render(msg, aaa.SessCreate(msg, share.Append(aaa.USERNAME)))
			case FIELD:
				msg.Option(ice.MSG_USERNAME, share.Append(aaa.USERNAME))
				msg.Option(ice.MSG_USERROLE, share.Append(aaa.USERROLE))
			}
		}
		return cmds, true // 白名单
	}

	if msg.Warn(msg.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, r.URL.Path) {
		msg.Render(STATUS, http.StatusUnauthorized, ice.ErrNotLogin)
		return cmds, false // 未登录
	}
	if msg.Warn(!msg.Right(r.URL.Path)) {
		msg.Render(STATUS, http.StatusForbidden, ice.ErrNotRight)
		return cmds, false // 未授权
	}
	return cmds, true
}

const (
	WEB_LOGIN = "_login"

	SSO = "sso"
)
const SERVE = "serve"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVE: {Name: SERVE, Help: "服务器", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
				tcp.LOCALHOST, true, aaa.BLACK, kit.Dict(), aaa.WHITE, kit.Dict(
					LOGIN, true, SPACE, true, SHARE, true,
					ice.VOLCANOS, true, ice.INTSHELL, true,
					ice.REQUIRE, true, ice.PUBLISH, true,
				), LOGHEADERS, false,

				kit.SSH_STATIC, kit.Dict("/", ice.USR_VOLCANOS),
				ice.VOLCANOS, kit.Dict(kit.MDB_PATH, ice.USR_VOLCANOS, kit.SSH_INDEX, "page/index.html",
					kit.SSH_REPOS, "https://github.com/shylinux/volcanos", kit.SSH_BRANCH, kit.SSH_MASTER,
				), ice.PUBLISH, ice.USR_PUBLISH,

				ice.INTSHELL, kit.Dict(kit.MDB_PATH, ice.USR_INTSHELL, kit.SSH_INDEX, ice.INDEX_SH,
					kit.SSH_REPOS, "https://github.com/shylinux/intshell", kit.SSH_BRANCH, kit.SSH_MASTER,
				), ice.REQUIRE, ".ish/pluged",
			)},
		},
		Commands: map[string]*ice.Command{
			SERVE: {Name: "serve name auto start", Help: "服务器", Action: map[string]*ice.Action{
				aaa.BLACK: {Name: "black", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range arg {
						m.Conf(SERVE, kit.Keys(kit.MDB_META, aaa.BLACK, k), true)
					}
				}},
				aaa.WHITE: {Name: "white", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range arg {
						m.Conf(SERVE, kit.Keys(kit.MDB_META, aaa.WHITE, k), true)
					}
				}},
				cli.START: {Name: "start dev= name=self proto=http host= port=9020", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if cli.NodeInfo(m, SERVER, ice.Info.HostName); m.Option(tcp.PORT) == tcp.RANDOM {
						m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
					}

					m.Target().Start(m, kit.MDB_NAME, m.Option(kit.MDB_NAME), tcp.HOST, m.Option(tcp.HOST), tcp.PORT, m.Option(tcp.PORT))
					m.Sleep(ice.MOD_TICK)

					m.Option(kit.MDB_NAME, "")
					for _, k := range kit.Split(m.Option(SPIDE_DEV)) {
						m.Cmd(SPACE, tcp.DIAL, SPIDE_DEV, k, kit.MDB_NAME, ice.Info.NodeName)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,status,name,port,dev")
				m.Cmdy(mdb.SELECT, SERVE, "", mdb.HASH, kit.MDB_NAME, arg)
			}},

			"/volcanos/": {Name: "/volcanos/", Help: "浏览器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.RenderDownload(path.Join(m.Conf(SERVE, kit.Keym(ice.VOLCANOS, kit.MDB_PATH)), path.Join(arg...)))
			}},
			"/intshell/": {Name: "/intshell/", Help: "命令行", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.RenderDownload(path.Join(m.Conf(SERVE, kit.Keym(ice.INTSHELL, kit.MDB_PATH)), path.Join(arg...)))
			}},
			"/publish/": {Name: "/publish/", Help: "私有云", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_local(m, m.Conf(SERVE, kit.Keym(ice.PUBLISH)), path.Join(arg...))
			}},
			"/require/": {Name: "/require/", Help: "公有云", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_repos(m, path.Join(arg[0], arg[1], arg[2]), arg[3:]...)
			}},
		}})
}
