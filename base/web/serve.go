package web

import (
	"net/http"
	"net/url"
	"path"
	"regexp"
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
)

func _serve_address(m *ice.Message) string { return Domain(tcp.LOCALHOST, m.Option(tcp.PORT)) }
func _serve_start(m *ice.Message) {
	defer kit.For(kit.Split(m.Option(ice.DEV)), func(v string) {
		m.Sleep("10ms").Cmd(SPACE, tcp.DIAL, ice.DEV, v, mdb.NAME, ice.Info.NodeName, TOKEN, m.Option(TOKEN))
	})
	kit.If(m.Option(aaa.USERNAME), func() { aaa.UserRoot(m, m.Option(aaa.USERNICK), m.Option(aaa.USERNAME)) })
	kit.If(m.Option(tcp.PORT) == tcp.RANDOM, func() { m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT)) })
	// kit.If(cli.IsWindows(), func() {
	m.Go(func() { m.Cmd(SPIDE, ice.OPS, _serve_address(m)+"/exit", ice.Maps{CLIENT_TIMEOUT: "30ms"}) })
	m.Sleep("30ms")
	// })
	cli.NodeInfo(m, kit.Select(ice.Info.Hostname, m.Option(tcp.NODENAME)), SERVER)
	m.Start("", m.OptionSimple(tcp.HOST, tcp.PORT)...)
}
func _serve_main(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	const (
		X_REAL_IP       = "X-Real-Ip"
		X_REAL_PORT     = "X-Real-Port"
		X_FORWARDED_FOR = "X-Forwarded-For"
		INDEX_MODULE    = "Index-Module"
	)
	if r.Header.Get(INDEX_MODULE) == "" {
		r.Header.Set(INDEX_MODULE, m.Prefix())
	} else {
		return true
	}
	if ip := r.Header.Get(X_REAL_IP); ip != "" {
		if r.Header.Set(ice.MSG_USERIP, ip); r.Header.Get(X_REAL_PORT) != "" {
			r.Header.Set(ice.MSG_USERADDR, ip+nfs.DF+r.Header.Get(X_REAL_PORT))
		}
	} else if ip := r.Header.Get(X_FORWARDED_FOR); ip != "" {
		r.Header.Set(ice.MSG_USERIP, kit.Split(ip)[0])
	} else if strings.HasPrefix(r.RemoteAddr, "[") {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, "]")[0][1:])
	} else {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, nfs.DF)[0])
	}
	if m.Logs(r.Header.Get(ice.MSG_USERIP), r.Method, r.URL.String()); r.Method == http.MethodGet {
		if msg := m.Spawn(w, r).Options(ice.MSG_USERUA, r.UserAgent()); path.Join(r.URL.Path) == nfs.PS {
			if !msg.IsCliUA() {
				if r.URL.Path = kit.Select(nfs.PS, mdb.Config(m, ice.MAIN)); path.Join(r.URL.Path) != nfs.PS {
					return true
				}
			}
			return !Render(RenderMain(msg), msg.Option(ice.MSG_OUTPUT), kit.List(msg.Optionv(ice.MSG_ARGS))...)
		} else if p := path.Join(kit.Select(ice.USR_VOLCANOS, ice.USR_INTSHELL, msg.IsCliUA()), r.URL.Path); nfs.Exists(msg, p) {
			return !Render(msg, ice.RENDER_DOWNLOAD, p)
		}
	} else if r.Method == http.MethodPost && path.Join(r.URL.Path) == nfs.PS {
		r.URL.Path = kit.Select(nfs.PS, mdb.Config(m, ice.MAIN))
	}
	return true
}
func _serve_handle(key string, cmd *ice.Command, m *ice.Message, w http.ResponseWriter, r *http.Request) {
	debug := strings.Contains(r.URL.String(), "debug=true") || strings.Contains(r.Header.Get(Referer), "debug=true")
	_log := func(level string, arg ...ice.Any) *ice.Message {
		if debug || arg[0] == ice.MSG_CMDS {
			return m.Logs(strings.Title(level), arg...)
		}
		return m
	}
	if u, e := url.Parse(r.Header.Get(Referer)); e == nil {
		add := func(k, v string) { _log(nfs.PATH, k, m.Option(k, v)) }
		switch arg := strings.Split(strings.TrimPrefix(u.Path, nfs.PS), nfs.PS); arg[0] {
		case CHAT:
			kit.For(arg[1:], func(k, v string) { add(k, v) })
		case SHARE:
			add(arg[0], arg[1])
		}
		kit.For(u.Query(), func(k string, v []string) { _log(ctx.ARGS, k, v).Optionv(k, v) })
	}
	kit.For(kit.ParseQuery(r.URL.RawQuery), func(k string, v []string) { m.Optionv(k, v) })
	switch r.Header.Get(ContentType) {
	case ApplicationJSON:
		kit.For(kit.UnMarshal(r.Body), func(k string, v ice.Any) { m.Optionv(k, v) })
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(ContentLength))))
		kit.For(r.PostForm, func(k string, v []string) { _log(FORM, k, kit.Join(v, lex.SP)).Optionv(k, v) })
	}
	m.Option(ice.MSG_COUNT, "0")
	kit.For(r.Cookies(), func(k, v string) { m.Optionv(k, v) })
	m.OptionDefault(ice.MSG_HEIGHT, "480", ice.MSG_WIDTH, "320")
	m.Options(ice.MSG_USERWEB, _serve_domain(m), ice.MSG_USERPOD, m.Option(ice.POD))
	m.Options(ice.MSG_USERUA, r.Header.Get(UserAgent), ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	m.Debug("what %v", m.Option(ice.MSG_USERWEB))
	m.Debug("what %v", CookieName(m.Option(ice.MSG_USERWEB)))

	m.Options(ice.MSG_SESSID, kit.Select(m.Option(ice.MSG_SESSID), m.Option(CookieName(m.Option(ice.MSG_USERWEB)))))
	kit.If(m.Optionv(ice.MSG_CMDS) == nil, func() {
		kit.If(strings.TrimPrefix(r.URL.Path, key), func(p string) { m.Optionv(ice.MSG_CMDS, strings.Split(p, nfs.PS)) })
	})
	defer func() { Render(m, m.Option(ice.MSG_OUTPUT), kit.List(m.Optionv(ice.MSG_ARGS))...) }()
	if cmds, ok := _serve_auth(m, key, kit.Simple(m.Optionv(ice.MSG_CMDS)), w, r); ok {
		defer func() { m.Cost(kit.Format("%s: %s %v", r.Method, m.PrefixPath()+path.Join(cmds...), m.FormatSize())) }()
		m.Option(ice.MSG_OPTS, kit.Simple(m.Optionv(ice.MSG_OPTION), func(k string) bool { return !strings.HasPrefix(k, ice.MSG_SESSID) }))
		if m.Detailv(m.PrefixKey(), cmds); len(cmds) > 1 && cmds[0] == ctx.ACTION {
			m.ActionHand(cmd, key, cmds[1], cmds[2:]...)
		} else {
			m.CmdHand(cmd, key, cmds...)
		}
	}
}
func _serve_domain(m *ice.Message) string {
	return kit.GetValid(
		func() string { return kit.Select("", m.R.Header.Get(Referer), m.R.Method == http.MethodPost) },
		func() string { return m.R.Header.Get("X-Host") },
		func() string { return ice.Info.Domain },
		func() string {
			if b, e := regexp.MatchString("^[0-9.]+$", m.R.Host); b && e == nil {
				return kit.Format("%s://%s:%s", kit.Select(HTTPS, HTTP, m.R.TLS == nil), m.R.Host, m.Option(tcp.PORT))
			}
			return kit.Format("%s://%s", kit.Select(HTTPS, HTTP, m.R.TLS == nil), m.R.Host)
		},
	)
}
func _serve_auth(m *ice.Message, key string, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	if r.URL.Path == PP(SPACE) {
		return cmds, true
	}
	defer func() { m.Options(ice.MSG_CMDS, "", ice.MSG_SESSID, "") }()
	if aaa.SessCheck(m, m.Option(ice.MSG_SESSID)); m.Option(ice.MSG_USERNAME) == "" && ice.Info.Localhost {
		ls := kit.Simple(mdb.Cache(m, m.Option(ice.MSG_USERIP), func() ice.Any {
			if tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
				aaa.UserRoot(m)
				return kit.Simple(m.Time(), m.OptionSplit(ice.MSG_USERNICK, ice.MSG_USERNAME, ice.MSG_USERROLE))
			}
			return nil
		}))
		if len(ls) > 0 {
			m.Auth(aaa.USERNICK, m.Option(ice.MSG_USERNICK, ls[1]), aaa.USERNAME, m.Option(ice.MSG_USERNAME, ls[2]), aaa.USERROLE, m.Option(ice.MSG_USERROLE, ls[3]), CACHE, ls[0])
		}
	}
	m.Cmd(COUNT, mdb.CREATE, aaa.IP, m.Option(ice.MSG_USERIP), m.Option(ice.MSG_USERUA), kit.Dict(ice.LOG_DISABLE, ice.TRUE))
	m.Cmd(COUNT, mdb.CREATE, m.R.Method, m.R.URL.Path, kit.Join(kit.Simple(m.Optionv(ice.MSG_CMDS)), " "), kit.Dict(ice.LOG_DISABLE, ice.TRUE))
	return cmds, aaa.Right(m, key, cmds)
}

const (
	SERVE_START = "serve.start"

	HTTP   = "http"
	HTTPS  = "https"
	DOMAIN = "domain"
	ORIGIN = "origin"
	FORM   = "form"
	BODY   = "body"

	ApplicationJSON  = "application/json"
	ApplicationOctet = "application/octet-stream"
)
const SERVE = "serve"

func init() {
	Index.MergeCommands(ice.Commands{
		"/exit": {Hand: func(m *ice.Message, arg ...string) { m.Cmd(ice.EXIT) }},
		SERVE: {Name: "serve name auto start dark system", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				cli.NodeInfo(m, ice.Info.Pathname, WORKER)
				gdb.Watch(m, SERVE_START)
				aaa.White(m, nfs.REQUIRE)
			}},
			DOMAIN: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(len(arg) > 0, func() { ice.Info.Domain, ice.Info.Localhost = arg[0], false })
				m.Echo(ice.Info.Domain)
			}},
			cli.START: {Name: "start dev proto host port=9020 nodename username usernick", Hand: func(m *ice.Message, arg ...string) {
				_serve_start(m)
			}},
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					m.Option(ice.MSG_USERIP, "127.0.0.1")
					cli.Opens(m, mdb.Config(m, cli.OPEN))
					ssh.PrintQRCode(m, tcp.PublishLocalhost(m, _serve_address(m)))
				})
			}},
			cli.SYSTEM: {Help: "系统", Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, "System Settings.app") }},
			"dark": {Help: "主题", Hand: func(m *ice.Message, arg ...string) {
				if !tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
					return
				}
				m.Cmd(cli.SYSTEM, "osascript", "-e", `tell app "System Events" to tell appearance preferences to set dark mode to not dark mode`)
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,status,name,proto,host,port"), mdb.ClearOnExitHashAction())},
	})
	ice.AddMergeAction(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) {
		if strings.HasPrefix(sub, nfs.PS) {
			kit.If(action.Hand == nil, func() { action.Hand = cmd.Hand })
			sub = kit.Select(P(key, sub), PP(key, sub), strings.HasSuffix(sub, nfs.PS))
			actions := ice.Actions{}
			for k, v := range cmd.Actions {
				if !kit.IsIn(k, ice.CTX_INIT, ice.CTX_EXIT) {
					actions[k] = v
				}
			}
			c.Commands[sub] = &ice.Command{Name: kit.Select(cmd.Name, action.Name), Actions: ice.MergeActions(actions, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
				msg := m.Spawn(c, key, cmd)
				defer m.Copy(msg)
				action.Hand(msg, arg...)
			}}
		}
	})
}
func Domain(host, port string) string { return kit.Format("%s://%s:%s", HTTP, host, port) }
func Script(m *ice.Message, str string, arg ...ice.Any) string {
	return ice.Render(m, ice.RENDER_SCRIPT, kit.Format(str, arg...))
}
func ChatCmdPath(arg ...string) string { return path.Join("/chat/cmd/", path.Join(arg...)) }
func RequireFile(m *ice.Message, file string) string {
	if strings.HasPrefix(file, nfs.PS) || strings.HasPrefix(file, ice.HTTP) {
		return file
	} else if file != "" {
		return "/require/" + file
	}
	return ""
}
