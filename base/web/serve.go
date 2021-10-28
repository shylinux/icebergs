package web

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
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
	if m.Config(LOGHEADERS) == ice.TRUE {
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

	// 代码管理
	if strings.HasPrefix(r.URL.Path, "/x/") {
		r.URL.Path = strings.Replace(r.URL.Path, "/x/", "/code/git/repos/", -1)
		return true
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
		Render(msg, ice.RENDER_DOWNLOAD, path.Join(m.Config(kit.Keys(repos, kit.MDB_PATH)), m.Config(kit.Keys(repos, kit.MDB_INDEX))))
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
		switch ls[2] {
		case "local":
		default:
			msg.Logs("refer", ls[1], ls[2])
			msg.Option(ls[1], ls[2])
		}
	case "chat":
		switch kit.Select("", ls, 2) {
		case "pod":
			msg.Logs("refer", ls[2], ls[3])
			msg.Option(ls[2], ls[3])
		}
	case "pod":
		msg.Logs("refer", ls[1], ls[2])
		msg.Option(ls[1], ls[2])
	}
}
func _serve_handle(key string, cmd *ice.Command, msg *ice.Message, w http.ResponseWriter, r *http.Request) {
	// 环境变量
	msg.Option(ice.MSG_OUTPUT, "")
	msg.Option(ice.MSG_SESSID, "")
	for _, v := range r.Cookies() {
		msg.Option(v.Name, v.Value)
	}

	// 请求变量
	if u, e := url.Parse(r.Header.Get("Referer")); e == nil {
		_serve_params(msg, u.Path)
		for k, v := range u.Query() {
			msg.Logs("refer", k, v)
			msg.Option(k, v)
		}
	}
	_serve_params(msg, r.URL.Path)

	// 请求地址
	msg.Option(ice.MSG_USERWEB, kit.Select(msg.Conf(SPACE, kit.Keym(kit.MDB_DOMAIN)), kit.Select(r.Header.Get("X-Host"), r.Header.Get("Referer"))))
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
		if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e, ice.ErrNotFound, data) {
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
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(ContentLength))))
		if r.ParseForm(); len(r.PostForm) > 0 {
			for k, v := range r.PostForm {
				msg.Logs("form", k, v)
			}
		}
	}

	// 请求参数
	for k, v := range r.Form {
		if msg.IsCliUA() {
			for i, p := range v {
				v[i], _ = url.QueryUnescape(p)
			}
		}
		if msg.Optionv(k, v); k == ice.MSG_SESSID {
			RenderCookie(msg, v[0])
		}
	}

	// 请求命令
	if msg.Option(ice.MSG_USERPOD, msg.Option(ice.POD)); msg.Optionv(ice.MSG_CMDS) == nil {
		if p := strings.TrimPrefix(r.URL.Path, key); p != "" {
			msg.Optionv(ice.MSG_CMDS, strings.Split(p, "/"))
		}
	}

	// 执行命令
	if cmds, ok := _serve_login(msg, key, kit.Simple(msg.Optionv(ice.MSG_CMDS)), w, r); ok {
		msg.Option(ice.MSG_OPTS, msg.Optionv(ice.MSG_OPTION))
		msg.Target().Cmd(msg, key, cmds...)
		msg.Cost(kit.Format("%s %v %v", r.URL.Path, cmds, msg.FormatSize()))
	}

	// 输出响应
	_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
	Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
}
func _serve_login(msg *ice.Message, key string, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	msg.Option(ice.MSG_USERROLE, aaa.VOID)
	msg.Option(ice.MSG_USERNAME, "")

	if msg.Option(ice.MSG_SESSID) != "" {
		aaa.SessCheck(msg, msg.Option(ice.MSG_SESSID))
		// 会话认证
	}

	if msg.Option(ice.MSG_USERNAME) == "" && msg.Config(tcp.LOCALHOST) == ice.TRUE && tcp.IsLocalHost(msg, msg.Option(ice.MSG_USERIP)) {
		aaa.UserRoot(msg)
		// 主机认证
	}

	if _, ok := msg.Target().Commands[WEB_LOGIN]; ok {
		// 权限检查
		msg.Target().Cmd(msg, WEB_LOGIN, kit.Simple(key, cmds)...)
		return cmds, msg.Result(0) != ice.ErrWarn && msg.Result(0) != ice.FALSE
	}

	if ls := strings.Split(r.URL.Path, "/"); msg.Config(kit.Keys(aaa.BLACK, ls[1])) == ice.TRUE {
		return cmds, false // 黑名单
	} else if msg.Config(kit.Keys(aaa.WHITE, ls[1])) == ice.TRUE {
		if msg.Option(ice.MSG_USERNAME) == "" && msg.Option(SHARE) != "" {
			switch share := msg.Cmd(SHARE, msg.Option(SHARE)); share.Append(kit.MDB_TYPE) {
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
	if !msg.Right(r.URL.Path) {
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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SERVE: {Name: SERVE, Help: "服务器", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,status,name,port,dev",
			tcp.LOCALHOST, ice.TRUE, aaa.BLACK, kit.Dict(), aaa.WHITE, kit.Dict(
				LOGIN, ice.TRUE, SHARE, ice.TRUE, SPACE, ice.TRUE,
				ice.VOLCANOS, ice.TRUE, ice.PUBLISH, ice.TRUE,
				ice.INTSHELL, ice.TRUE, ice.REQUIRE, ice.TRUE,
				"x", ice.TRUE,
			), LOGHEADERS, ice.FALSE,

			kit.MDB_PATH, kit.Dict("/", ice.USR_VOLCANOS),
			ice.VOLCANOS, kit.Dict(kit.MDB_PATH, ice.USR_VOLCANOS, kit.MDB_INDEX, "page/index.html",
				kit.SSH_REPOS, "https://shylinux.com/x/volcanos", kit.SSH_BRANCH, kit.SSH_MASTER,
			), ice.PUBLISH, ice.USR_PUBLISH,

			ice.INTSHELL, kit.Dict(kit.MDB_PATH, ice.USR_INTSHELL, kit.MDB_INDEX, ice.INDEX_SH,
				kit.SSH_REPOS, "https://shylinux.com/x/intshell", kit.SSH_BRANCH, kit.SSH_MASTER,
			), ice.REQUIRE, ".ish/pluged",
		)},
	}, Commands: map[string]*ice.Command{
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(SERVE).Table(func(index int, value map[string]string, head []string) {
				m.Done(value[kit.MDB_STATUS] == tcp.START)
			})
		}},
		SERVE: {Name: "serve name auto start", Help: "服务器", Action: ice.MergeAction(map[string]*ice.Action{
			aaa.BLACK: {Name: "black", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
				for _, k := range arg {
					m.Log_CREATE(aaa.BLACK, k)
					m.Config(kit.Keys(aaa.BLACK, k), ice.TRUE)
				}
			}},
			aaa.WHITE: {Name: "white", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
				for _, k := range arg {
					m.Log_CREATE(aaa.WHITE, k)
					m.Config(kit.Keys(aaa.WHITE, k), ice.TRUE)
				}
			}},
			cli.START: {Name: "start dev name=ops proto=http host port=9020", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				if cli.NodeInfo(m, SERVER, ice.Info.HostName); m.Option(tcp.PORT) == tcp.RANDOM {
					m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
				}

				m.Target().Start(m, m.OptionSimple(kit.MDB_NAME, tcp.HOST, tcp.PORT)...)
				m.Sleep(ice.MOD_TICK)

				m.Option(kit.MDB_NAME, "")
				for _, k := range kit.Split(m.Option(ice.DEV)) {
					m.Cmd(SPACE, tcp.DIAL, ice.DEV, k, kit.MDB_NAME, ice.Info.NodeName)
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},

		"/intshell/": {Name: "/intshell/", Help: "命令行", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.RenderIndex(SERVE, ice.INTSHELL, arg...)
		}},
		"/volcanos/": {Name: "/volcanos/", Help: "浏览器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.RenderIndex(SERVE, ice.VOLCANOS, arg...)
		}},
		"/require/": {Name: "/require/", Help: "代码库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_share_repos(m, path.Join(arg[0], arg[1], arg[2]), arg[3:]...)
		}},
		"/publish/": {Name: "/publish/", Help: "定制化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if arg[0] == ice.ORDER_JS {
				if p := path.Join(ice.USR_PUBLISH, ice.ORDER_JS); m.PodCmd(nfs.CAT, p) {
					m.RenderResult()
					return
				}
			}
			_share_local(m, m.Conf(SERVE, kit.Keym(ice.PUBLISH)), path.Join(arg...))
		}},
		"/help/": {Name: "/help/", Help: "帮助", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "tutor.shy")
			}
			if len(arg) > 0 && arg[0] != ctx.ACTION {
				arg[0] = "src/help/" + arg[0]
			}
			m.Cmdy("web.chat./cmd/", arg)
		}},
	}})
}
