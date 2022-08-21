package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

var rewriteList = []ice.Any{}

func AddRewrite(cb ice.Any) { rewriteList = append(rewriteList, cb) }

func _serve_rewrite(m *ice.Message) {
	AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != SPIDE_GET {
			return false
		}

		msg, repos := m.Spawn(SERVE, w, r), kit.Select(ice.INTSHELL, ice.VOLCANOS, strings.Contains(r.Header.Get(UserAgent), "Mozilla/5.0"))
		switch r.URL.Path {
		case ice.PS:
			if repos == ice.VOLCANOS {
				if s := msg.Cmdx("web.chat.website", lex.PARSE, ice.INDEX_IML, "Header", "", "River", "", "Footer", ""); s != "" {
					Render(msg, ice.RENDER_RESULT, s)
					return true // 定制主页
				}
			}
			Render(msg, ice.RENDER_DOWNLOAD, path.Join(msg.Config(kit.Keys(repos, nfs.PATH)), msg.Config(kit.Keys(repos, INDEX))))
			return true // 默认主页

		case PP(ice.HELP):
			r.URL.Path = P(ice.HELP, ice.TUTOR_SHY)
		}

		p := path.Join(ice.USR, repos, r.URL.Path)
		if _, e := nfs.DiskFile.StatFile(p); e == nil {
			http.ServeFile(w, r, kit.Path(p))
			return true
		} else if f, e := nfs.PackFile.OpenFile(p); e == nil {
			defer f.Close()
			RenderType(w, p, "")
			io.Copy(w, f)
			return true
		}
		return false
	})
}
func _serve_domain(m *ice.Message) string {
	if p := ice.Info.Domain; p != "" {
		return p
	}
	if p := m.R.Header.Get("X-Host"); p != "" {
		return p
	}
	if m.R.Method == SPIDE_POST {
		if p := m.R.Header.Get(Referer); p != "" {
			return p
		}
	}
	if m.R.TLS == nil {
		return kit.Format("http://%s", m.R.Host)
	} else {
		return kit.Format("https://%s", m.R.Host)
	}
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
func _serve_start(m *ice.Message) {
	ice.Info.Domain = kit.Select(kit.Format("%s://%s:%s", m.Option(tcp.PROTO), kit.Select(m.Cmd(tcp.HOST).Append(aaa.IP), m.Option(tcp.HOST)), m.Option(tcp.PORT)), ice.Info.Domain)
	if cli.NodeInfo(m, SERVER, kit.Select(ice.Info.HostName, m.Option("nodename"))); m.Option(tcp.PORT) == tcp.RANDOM {
		m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
	}

	if m.Option("staffname") != "" {
		m.Config("staffname", m.Option(aaa.USERNAME, m.Option("staffname")))
	}
	aaa.UserRoot(m, m.Option(aaa.PASSWORD), m.Option(aaa.USERNAME), m.Option(aaa.USERROLE))

	m.Target().Start(m, m.OptionSimple(tcp.HOST, tcp.PORT)...)
	m.Go(func() { m.Cmd(BROAD, SERVE) })
	m.Sleep300ms()

	for _, k := range kit.Split(m.Option(ice.DEV)) {
		m.Cmd(SPACE, tcp.DIAL, ice.DEV, k, mdb.NAME, ice.Info.NodeName)
	}
}

func _serve_main(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Index-Module") == "" {
		r.Header.Set("Index-Module", m.Prefix())
	} else {
		return true
	}

	// 用户地址
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		if r.Header.Set(ice.MSG_USERIP, ip); r.Header.Get("X-Real-Port") != "" {
			r.Header.Set(ice.MSG_USERADDR, ip+":"+r.Header.Get("X-Real-Port"))
		}
	} else if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		r.Header.Set(ice.MSG_USERIP, kit.Split(ip)[0])
	} else if strings.HasPrefix(r.RemoteAddr, "[") {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, "]")[0][1:])
	} else {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, ":")[0])
	}
	meta := logs.FileLineMeta("")
	m.Info("%s %s %s", r.Header.Get(ice.MSG_USERIP), r.Method, r.URL, meta)

	// 参数日志
	if m.Config(LOGHEADERS) == ice.TRUE {
		for k, v := range r.Header {
			m.Info("%s: %v", k, kit.Format(v), meta)
		}
		m.Info("", meta)

		defer func() {
			m.Info("", meta)
			for k, v := range w.Header() {
				m.Info("%s: %v", k, kit.Format(v), meta)
			}
		}()
	}

	// 模块回调
	for _, h := range rewriteList {
		if m.Config(LOGHEADERS) == ice.TRUE {
			m.Info("%s: %v", r.URL.Path, kit.FileLine(h, 3), meta)
		}
		switch h := h.(type) {
		case func(w http.ResponseWriter, r *http.Request) func():
			defer h(w, r)

		case func(p string, w http.ResponseWriter, r *http.Request) bool:
			if h(r.URL.Path, w, r) {
				return false
			}
		case func(w http.ResponseWriter, r *http.Request) bool:
			if h(w, r) {
				return false
			}
		default:
			m.ErrorNotImplement(h)
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
	case ice.POD:
		msg.Logs("refer", ls[1], ls[2])
		msg.Option(ls[1], ls[2])
	case "chat":
		switch kit.Select("", ls, 2) {
		case ice.POD:
			msg.Logs("refer", ls[2], ls[3])
			msg.Option(ls[2], ls[3])
		}
	}
}
func _serve_handle(key string, cmd *ice.Command, msg *ice.Message, w http.ResponseWriter, r *http.Request) {
	// 地址参数
	if u, e := url.Parse(r.Header.Get(Referer)); e == nil {
		_serve_params(msg, u.Path)
		for k, v := range u.Query() {
			msg.Logs("refer", k, v)
			msg.Option(k, v)
		}
	}
	_serve_params(msg, r.URL.Path)

	// 解析参数
	switch r.Header.Get(ContentType) {
	case ContentJSON:
		defer r.Body.Close()
		var data ice.Any
		if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e, ice.ErrNotFound, data) {
			msg.Logs(mdb.IMPORT, mdb.VALUE, kit.Format(data))
			msg.Optionv(ice.MSG_USERDATA, data)
		}
		kit.Fetch(data, func(key string, value ice.Any) { msg.Optionv(key, value) })

	default:
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(ContentLength))))
		if r.ParseForm(); len(r.PostForm) > 0 {
			meta := logs.FileLineMeta("")
			for k, v := range r.PostForm {
				if len(v) > 1 {
					msg.Logs("form", k, len(v), kit.Join(v, ice.SP), meta)
				} else {
					msg.Logs("form", k, v, meta)
				}
			}
		}
	}

	// 请求参数
	msg.R, msg.W = r, w
	for k, v := range r.Form {
		if msg.IsCliUA() {
			for i, p := range v {
				v[i], _ = url.QueryUnescape(p)
			}
		}
		msg.Optionv(k, v)
	}
	for k, v := range r.PostForm {
		msg.Optionv(k, v)
	}
	for _, v := range r.Cookies() {
		msg.Option(v.Name, v.Value)
	}

	// 用户参数
	msg.Option(ice.MSG_USERWEB, _serve_domain(msg))
	msg.Option(ice.MSG_USERADDR, kit.Select(r.RemoteAddr, r.Header.Get(ice.MSG_USERADDR)))
	msg.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	msg.Option(ice.MSG_USERUA, r.Header.Get(UserAgent))
	if msg.Option(ice.POD) != "" {
		msg.Option(ice.MSG_USERPOD, msg.Option(ice.POD))
	}

	// 会话参数
	if sessid := msg.Option(CookieName(msg.Option(ice.MSG_USERWEB))); msg.Option(ice.MSG_SESSID) == "" {
		msg.Option(ice.MSG_SESSID, sessid)
	}

	// 解析命令
	if msg.Optionv(ice.MSG_CMDS) == nil {
		if p := strings.TrimPrefix(r.URL.Path, key); p != "" {
			msg.Optionv(ice.MSG_CMDS, strings.Split(p, ice.PS))
		}
	}

	// 执行命令
	if cmds, ok := _serve_login(msg, key, kit.Simple(msg.Optionv(ice.MSG_CMDS)), w, r); ok {
		defer func() { msg.Cost(kit.Format("%s %v %v", r.URL.Path, cmds, msg.FormatSize())) }()
		msg.Option(ice.MSG_OPTS, kit.Filter(kit.Simple(msg.Optionv(ice.MSG_OPTION)), func(k string) bool {
			return !strings.HasPrefix(k, ice.MSG_SESSID)
		}))
		msg.Target().Cmd(msg, key, cmds...)
	}

	// 输出响应
	switch args := msg.Optionv(ice.MSG_ARGS).(type) {
	case []ice.Any:
		Render(msg, msg.Option(ice.MSG_OUTPUT), args...)
	default:
		Render(msg, msg.Option(ice.MSG_OUTPUT), args)
	}
}
func _serve_login(msg *ice.Message, key string, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	aaa.SessCheck(msg, msg.Option(ice.MSG_SESSID)) // 会话认证

	if msg.Config("staffname") != "" {
		aaa.UserLogin(msg, r.Header.Get("Staffname"), "")
	}

	if msg.Option(ice.MSG_USERNAME) == "" && msg.Config(tcp.LOCALHOST) == ice.TRUE && tcp.IsLocalHost(msg, msg.Option(ice.MSG_USERIP)) {
		aaa.UserRoot(msg) // 本机认证
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
		return cmds, !msg.IsErr() && msg.Result(0) != ice.FALSE
	}

	if aaa.Right(msg, key, cmds) {
		return cmds, true
	}

	if msg.Warn(msg.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, r.URL.Path) {
		msg.Render(STATUS, http.StatusUnauthorized, ice.ErrNotLogin)
		return cmds, false // 未登录
	} else if !aaa.Right(msg, r.URL.Path) {
		msg.Render(STATUS, http.StatusForbidden, ice.ErrNotRight)
		return cmds, false // 未授权
	}
	return cmds, true
}

const (
	WEB_LOGIN = "_login"
	SSO       = "sso"

	DOMAIN = "domain"
	INDEX  = "index"
)
const SERVE = "serve"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		SERVE: {Name: SERVE, Help: "服务器", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,status,name,proto,host,port,dev",
			tcp.LOCALHOST, ice.TRUE, LOGHEADERS, ice.FALSE,
			nfs.PATH, kit.Dict(ice.PS, ice.USR_VOLCANOS),
			ice.VOLCANOS, kit.Dict(nfs.PATH, ice.USR_VOLCANOS, INDEX, "page/index.html",
				nfs.REPOS, "https://shylinux.com/x/volcanos", nfs.BRANCH, nfs.MASTER,
			),
			ice.INTSHELL, kit.Dict(nfs.PATH, ice.USR_INTSHELL, INDEX, ice.INDEX_SH,
				nfs.REPOS, "https://shylinux.com/x/intshell", nfs.BRANCH, nfs.MASTER,
			),
		)},
	}, Commands: ice.Commands{
		SERVE: {Name: "serve name auto start spide", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				cli.NodeInfo(m, WORKER, ice.Info.PathName)
				for _, p := range []string{LOGIN, SHARE, SPACE, ice.VOLCANOS, ice.INTSHELL, ice.PUBLISH, ice.REQUIRE, ice.HELP, ice.CMD} {
					m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, p)
				}
				_serve_rewrite(m)
				gdb.Watch(m, ssh.SOURCE_STDIO)
			}},
			ssh.SOURCE_STDIO: {Name: "source.stdio", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					m.Sleep("2s")
					m.Cmd(ssh.PRINTF, kit.Dict(nfs.CONTENT, ice.Render(m, ice.RENDER_QRCODE, m.Cmdx(SPACE, DOMAIN))+ice.NL))
				})
			}},
			DOMAIN: {Name: "domain", Help: "域名", Hand: func(m *ice.Message, arg ...string) {
				m.Config(tcp.LOCALHOST, ice.FALSE)
				ice.Info.Domain = arg[0]
			}},
			SPIDE: {Name: "spide", Help: "架构图", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 { // 模块列表
					_serve_spide(m, ice.PS, m.Target())
					ctx.DisplayStorySpide(m, lex.PREFIX, m.ActionKey(), nfs.ROOT, MergeLink(m, ice.PS))
				}
			}},
			cli.START: {Name: "start dev proto=http host port=9020 nodename password username userrole staffname", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				_serve_start(m)
			}},
		}, mdb.HashAction())},

		PP(ice.INTSHELL): {Name: "/intshell/", Help: "命令行", Hand: func(m *ice.Message, arg ...string) {
			RenderIndex(m, ice.INTSHELL, arg...)
		}},
		PP(ice.VOLCANOS): {Name: "/volcanos/", Help: "浏览器", Hand: func(m *ice.Message, arg ...string) {
			RenderIndex(m, ice.VOLCANOS, arg...)
		}},
		PP(ice.PUBLISH): {Name: "/publish/", Help: "定制化", Hand: func(m *ice.Message, arg ...string) {
			_share_local(aaa.UserRoot(m), ice.USR_PUBLISH, path.Join(arg...))
		}},
		PP(ice.REQUIRE): {Name: "/require/shylinux.com/x/volcanos/proto.js", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			_share_repos(m, path.Join(arg[0], arg[1], arg[2]), arg[3:]...)
		}},
		PP(ice.REQUIRE, ice.NODE_MODULES): {Name: "/require/node_modules/", Help: "依赖库", Hand: func(m *ice.Message, arg ...string) {
			p := path.Join(ice.SRC, ice.NODE_MODULES, path.Join(arg...))
			if !nfs.ExistsFile(m, p) {
				m.Cmd(cli.SYSTEM, "npm", "install", arg[0], kit.Dict(cli.CMD_DIR, path.Join(ice.SRC)))
			}
			m.RenderDownload(p)
		}},
		PP(ice.REQUIRE, ice.USR): {Name: "/require/usr/", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			_share_local(aaa.UserRoot(m), ice.USR, path.Join(arg...))
		}},
		PP(ice.REQUIRE, ice.SRC): {Name: "/require/src/", Help: "源代码", Hand: func(m *ice.Message, arg ...string) {
			_share_local(aaa.UserRoot(m), ice.SRC, path.Join(arg...))
		}},
		PP(ice.HELP): {Name: "/help/", Help: "帮助", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, ice.TUTOR_SHY)
			}
			if len(arg) > 0 && arg[0] != ctx.ACTION {
				arg[0] = path.Join(ice.SRC_HELP, arg[0])
			}
			m.Cmdy("web.chat./cmd/", arg)
		}},
	}})
}
