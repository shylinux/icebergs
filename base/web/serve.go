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

var rewriteList = []interface{}{}

func AddRewrite(cb interface{}) { rewriteList = append(rewriteList, cb) }

func _serve_main(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Index-Module") == "" {
		r.Header.Set("Index-Module", m.Prefix())
	} else {
		return true
	}

	// 用户地址
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		r.Header.Set(ice.MSG_USERIP, ip)
	} else if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		if r.Header.Set(ice.MSG_USERIP, ip); r.Header.Get("X-Real-Port") != "" {
			r.Header.Set(ice.MSG_USERADDR, ip+":"+r.Header.Get("X-Real-Port"))
		}
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
			m.Info("")
			for k, v := range w.Header() {
				m.Info("%s: %v", k, kit.Format(v))
			}
		}()
	}

	// 模块回调
	for _, h := range rewriteList {
		if m.Config(LOGHEADERS) == ice.TRUE {
			m.Info("%s: %v", r.URL.Path, kit.FileLine(h, 3))
		}
		switch h := h.(type) {
		case func(w http.ResponseWriter, r *http.Request) func():
			defer h(w, r)

		case func(w http.ResponseWriter, r *http.Request) bool:
			if h(w, r) {
				return false
			}
		}
	}
	return true
}
func _serve_params(msg *ice.Message, path string) {
	switch ls := strings.Split(path, ice.PS); kit.Select("", ls, 1) {
	case SHARE:
		switch ls[2] {
		case "local":
		default:
			msg.Logs("refer", ls[1], ls[2])
			msg.Option(ls[1], ls[2])
		}
	case "chat":
		switch kit.Select("", ls, 2) {
		case ice.POD:
			msg.Logs("refer", ls[2], ls[3])
			msg.Option(ls[2], ls[3])
		}
	case ice.POD:
		msg.Logs("refer", ls[1], ls[2])
		msg.Option(ls[1], ls[2])
	}
}
func _serve_handle(key string, cmd *ice.Command, msg *ice.Message, w http.ResponseWriter, r *http.Request) {
	// 会话变量
	msg.Option(ice.MSG_SESSID, "")
	for _, v := range r.Cookies() {
		msg.Option(v.Name, v.Value)
	}

	// 请求参数
	if u, e := url.Parse(r.Header.Get("Referer")); e == nil {
		_serve_params(msg, u.Path)
		for k, v := range u.Query() {
			msg.Logs("refer", k, v)
			msg.Option(k, v)
		}
	}
	_serve_params(msg, r.URL.Path)

	// 请求数据
	switch r.Header.Get(ContentType) {
	case ContentJSON:
		var data interface{}
		if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e, ice.ErrNotFound, data) {
			msg.Log_IMPORT(mdb.VALUE, kit.Format(data))
			msg.Optionv(ice.MSG_USERDATA, data)
		}
		kit.Fetch(data, func(key string, value interface{}) { msg.Optionv(key, value) })

	default:
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(ContentLength))))
		if r.ParseForm(); len(r.PostForm) > 0 {
			for k, v := range r.PostForm {
				if len(v) > 1 {
					msg.Logs("form", k, len(v), v)
				} else {
					msg.Logs("form", k, v)
				}
			}
		}
	}

	// 请求地址
	msg.Option(ice.MSG_USERWEB, kit.Select(msg.Config(DOMAIN), kit.Select(r.Header.Get("Referer"), r.Header.Get("X-Host"))))
	msg.Option(ice.MSG_USERADDR, kit.Select(r.RemoteAddr, r.Header.Get(ice.MSG_USERADDR)))
	msg.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	msg.Option(ice.MSG_USERUA, r.Header.Get("User-Agent"))
	msg.R, msg.W = r, w

	// 会话别名
	if sessid := msg.Option(CookieName(msg.Option(ice.MSG_USERWEB))); sessid != "" {
		msg.Option(ice.MSG_SESSID, sessid)
	}

	// 参数转换
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

	if msg.Option(ice.MSG_USERPOD, msg.Option(ice.POD)); msg.Optionv(ice.MSG_CMDS) == nil {
		if p := strings.TrimPrefix(r.URL.Path, key); p != "" { // 地址命令
			msg.Optionv(ice.MSG_CMDS, strings.Split(p, ice.PS))
		}
	}

	msg.Option(ice.MSG_OUTPUT, "")
	msg.Option(ice.MSG_ARGS, kit.List())
	if cmds, ok := _serve_login(msg, key, kit.Simple(msg.Optionv(ice.MSG_CMDS)), w, r); ok {
		msg.Option(ice.MSG_OPTS, msg.Optionv(ice.MSG_OPTION))
		msg.Target().Cmd(msg, key, cmds...) // 执行命令
		msg.Cost(kit.Format("%s %v %v", r.URL.Path, cmds, msg.FormatSize()))
	}

	// 输出响应
	Render(msg, msg.Option(ice.MSG_OUTPUT), msg.Optionv(ice.MSG_ARGS).([]interface{})...)
}
func _serve_login(msg *ice.Message, key string, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	aaa.SessCheck(msg, msg.Option(ice.MSG_SESSID)) // 会话认证

	if msg.Option(ice.MSG_USERNAME) == "" && msg.Config(tcp.LOCALHOST) == ice.TRUE && tcp.IsLocalHost(msg, msg.Option(ice.MSG_USERIP)) {
		aaa.UserRoot(msg) // 主机认证
	}

	if msg.Option(ice.MSG_USERNAME) == "" && msg.Option(SHARE) != "" {
		switch share := msg.Cmd(SHARE, msg.Option(SHARE)); share.Append(mdb.TYPE) {
		case STORM, FIELD: // 共享认证
			msg.Option(ice.MSG_USERNAME, share.Append(aaa.USERNAME))
			msg.Option(ice.MSG_USERROLE, share.Append(aaa.USERROLE))
			msg.Option(ice.MSG_USERNICK, share.Append(aaa.USERNICK))
		}
	}

	if _, ok := msg.Target().Commands[WEB_LOGIN]; ok { // 单点认证
		msg.Target().Cmd(msg, WEB_LOGIN, kit.Simple(key, cmds)...)
		return cmds, msg.Result(0) != ice.ErrWarn && msg.Result(0) != ice.FALSE
	}

	if ls := strings.Split(r.URL.Path, ice.PS); msg.Config(kit.Keys(aaa.BLACK, ls[1])) == ice.TRUE {
		return cmds, false // 黑名单
	} else if msg.Config(kit.Keys(aaa.WHITE, ls[1])) == ice.TRUE {
		return cmds, true // 白名单
	}

	if msg.Warn(msg.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, r.URL.Path) {
		msg.Render(STATUS, http.StatusUnauthorized, ice.ErrNotLogin)
		return cmds, false // 未登录
	} else if !msg.Right(r.URL.Path) {
		msg.Render(STATUS, http.StatusForbidden, ice.ErrNotRight)
		return cmds, false // 未授权
	}
	return cmds, true
}
func _serve_spide(m *ice.Message, prefix string, c *ice.Context) {
	for k := range c.Commands {
		if strings.HasPrefix(k, ice.PS) {
			m.Push(nfs.PATH, path.Join(prefix, k)+kit.Select("", ice.PS, strings.HasSuffix(k, ice.PS)))
		}
	}
	for k, v := range c.Contexts {
		_serve_spide(m, path.Join(prefix, k), v)
	}
}

const (
	WEB_LOGIN = "_login"
	SSO       = "sso"

	DOMAIN = "domain"
	INDEX  = "index"
)
const SERVE = "serve"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SERVE: {Name: SERVE, Help: "服务器", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,status,name,port,dev",
			tcp.LOCALHOST, ice.TRUE, aaa.BLACK, kit.Dict(), aaa.WHITE, kit.Dict(
				LOGIN, ice.TRUE, SHARE, ice.TRUE, SPACE, ice.TRUE,
				ice.VOLCANOS, ice.TRUE, ice.PUBLISH, ice.TRUE,
				ice.INTSHELL, ice.TRUE, ice.REQUIRE, ice.TRUE,
				"cmd", ice.TRUE, ice.HELP, ice.TRUE,
			), LOGHEADERS, ice.FALSE,

			DOMAIN, "", nfs.PATH, kit.Dict(ice.PS, ice.USR_VOLCANOS),
			ice.VOLCANOS, kit.Dict(nfs.PATH, ice.USR_VOLCANOS, INDEX, "page/index.html",
				nfs.REPOS, "https://shylinux.com/x/volcanos", nfs.BRANCH, nfs.MASTER,
			), ice.PUBLISH, ice.USR_PUBLISH,

			ice.INTSHELL, kit.Dict(nfs.PATH, ice.USR_INTSHELL, INDEX, ice.INDEX_SH,
				nfs.REPOS, "https://shylinux.com/x/intshell", nfs.BRANCH, nfs.MASTER,
			), ice.REQUIRE, ".ish/pluged",
		)},
	}, Commands: map[string]*ice.Command{
		SERVE: {Name: "serve name auto start spide", Help: "服务器", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				cli.NodeInfo(m, WORKER, ice.Info.PathName)
				AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
					if r.Method == SPIDE_GET && r.URL.Path == ice.PS {
						msg := m.Spawn(SERVE, w, r)
						repos := kit.Select(ice.INTSHELL, ice.VOLCANOS, strings.Contains(r.Header.Get("User-Agent"), "Mozilla/5.0"))
						Render(msg, ice.RENDER_DOWNLOAD, path.Join(msg.Config(kit.Keys(repos, nfs.PATH)), msg.Config(kit.Keys(repos, INDEX))))
						return true // 网站主页
					}
					return false
				})
			}},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SERVE).Table(func(index int, value map[string]string, head []string) {
					m.Done(value[cli.STATUS] == tcp.START)
				})
			}},
			DOMAIN: {Name: "domain", Help: "域名", Hand: func(m *ice.Message, arg ...string) {
				ice.Info.Domain = m.Conf(SHARE, kit.Keym(DOMAIN, m.Config(DOMAIN, arg[0])))
			}},
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
			cli.START: {Name: "start dev name=ops proto=http host port=9020 nodename password username userrole", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				ice.Info.Domain = kit.Select(kit.Format("%s://%s:%s", m.Option(tcp.PROTO),
					kit.Select(m.Cmd(tcp.HOST).Append(aaa.IP), m.Option(tcp.HOST)), m.Option(tcp.PORT)), ice.Info.Domain)
				if cli.NodeInfo(m, SERVER, kit.Select(ice.Info.HostName, m.Option("nodename"))); m.Option(tcp.PORT) == tcp.RANDOM {
					m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
				}
				aaa.UserRoot(m, m.Option(aaa.PASSWORD), m.Option(aaa.USERNAME), m.Option(aaa.USERROLE))

				m.Target().Start(m, m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT)...)
				m.Sleep300ms()

				m.Option(mdb.NAME, "")
				for _, k := range kit.Split(m.Option(ice.DEV)) {
					m.Cmd(SPACE, tcp.DIAL, ice.DEV, k, mdb.NAME, ice.Info.NodeName)
				}
			}},
			"spide": {Name: "spide", Help: "架构图", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 { // 模块列表
					m.Display("/plugin/story/spide.js", "root", ice.ICE, "prefix", "spide")
					_serve_spide(m, ice.PS, m.Target())
					m.StatusTimeCount()
					return
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
