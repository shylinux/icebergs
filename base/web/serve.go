package web

import (
	"net/http"
	"net/url"
	"path"
	"regexp"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _serve_address(m *ice.Message) string { return Domain(tcp.LOCALHOST, m.Option(tcp.PORT)) }
func _serve_start(m *ice.Message) {
	kit.If(m.Option(aaa.USERNAME), func() { aaa.UserRoot(m, m.Option(aaa.USERNICK), m.Option(aaa.USERNAME)) })
	kit.If(m.Option(tcp.PORT) == tcp.RANDOM, func() { m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT)) })
	kit.If(runtime.GOOS == cli.WINDOWS || m.Cmdx(cli.SYSTEM, "lsof", "-i", ":"+m.Option(tcp.PORT)) != "", func() {
		m.Go(func() { m.Cmd(SPIDE, ice.OPS, _serve_address(m)+"/exit", ice.Maps{CLIENT_TIMEOUT: "300ms"}) }).Sleep300ms()
	})
	cli.NodeInfo(m, kit.Select(ice.Info.Hostname, m.Option(tcp.NODENAME)), SERVER)
	m.Start("", m.OptionSimple(tcp.HOST, tcp.PORT)...)
	kit.For(kit.Split(m.Option(ice.DEV)), func(dev string) {
		m.Sleep30ms(SPACE, tcp.DIAL, ice.DEV, dev, mdb.NAME, ice.Info.NodeName, m.OptionSimple(TOKEN))
	})
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
	if !kit.HasPrefix(r.URL.String(), VOLCANOS, REQUIRE_MODULES, INTSHELL) {
		r.Header.Set(ice.LOG_TRACEID, log.Traceid())
		m.Logs(r.Header.Get(ice.MSG_USERIP), r.Method, r.URL.String(), logs.TraceidMeta(r.Header.Get(ice.LOG_TRACEID)))
	}
	if path.Join(r.URL.Path) == nfs.PS && strings.HasPrefix(r.UserAgent(), Mozilla) {
		r.URL.Path = kit.Select(nfs.PS, mdb.Config(m, ice.MAIN))
	}
	if r.Method == http.MethodGet {
		msg := m.Spawn(w, r).Options(ice.MSG_USERUA, r.UserAgent(), ice.LOG_TRACEID, r.Header.Get(ice.LOG_TRACEID),
			ParseLink(m, kit.Select(r.URL.String(), r.Referer())))
		if path.Join(r.URL.Path) == nfs.PS {
			if Render(RenderMain(msg), msg.Option(ice.MSG_OUTPUT), kit.List(msg.Optionv(ice.MSG_ARGS))...) {
				return false
			}
		} else if _serve_static(msg, w, r) {
			return false
		}
	}
	return true
}
func _serve_static(msg *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if p := path.Join(kit.Select(ice.USR_VOLCANOS, ice.USR_INTSHELL, msg.IsCliUA()), r.URL.Path); nfs.Exists(msg, p) {
		return Render(msg, ice.RENDER_DOWNLOAD, p)
	} else if p = path.Join(nfs.USR, r.URL.Path); kit.HasPrefix(r.URL.Path, nfs.VOLCANOS, nfs.INTSHELL) && nfs.Exists(msg, p) {
		return Render(msg, ice.RENDER_DOWNLOAD, p)
	} else if p = strings.TrimPrefix(r.URL.Path, nfs.REQUIRE); kit.HasPrefix(r.URL.Path, ice.REQUIRE_SRC, nfs.REQUIRE+ice.USR_ICONS, nfs.REQUIRE+ice.USR_ICEBERGS) && nfs.Exists(msg, p) {
		ispod := kit.Contains(r.URL.String(), CHAT_POD, "pod=") || kit.Contains(r.Header.Get(Referer), CHAT_POD, "pod=")
		return !ispod && Render(msg, ice.RENDER_DOWNLOAD, p)
	} else if p = path.Join(ice.USR_MODULES, strings.TrimPrefix(r.URL.Path, ice.REQUIRE_MODULES)); kit.HasPrefix(r.URL.Path, ice.REQUIRE_MODULES) && nfs.Exists(msg, p) {
		return Render(msg, ice.RENDER_DOWNLOAD, p)
	} else {
		return false
	}
}
func _serve_handle(key string, cmd *ice.Command, m *ice.Message, w http.ResponseWriter, r *http.Request) {
	debug := strings.Contains(r.URL.String(), "debug=true") || strings.Contains(r.Header.Get(Referer), "debug=true")
	m.Option(ice.LOG_TRACEID, r.Header.Get(ice.LOG_TRACEID))
	_log := func(level string, arg ...ice.Any) *ice.Message {
		if debug || arg[0] == ice.MSG_CMDS {
			return m.Logs(strings.Title(level), arg...)
		}
		return m
	}
	_log("page", Referer, r.Header.Get(Referer))
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
	if r.Method == http.MethodGet && m.Option(ice.MSG_CMDS) != "" {
		_log(ctx.ARGS, ice.MSG_CMDS, m.Optionv(ice.MSG_CMDS))
	}
	switch r.Header.Get(ContentType) {
	case ApplicationJSON:
		kit.For(kit.UnMarshal(r.Body), func(k string, v ice.Any) { m.Optionv(k, v) })
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(ContentLength))))
		kit.For(r.PostForm, func(k string, v []string) { _log(FORM, k, kit.Join(v, lex.SP)).Optionv(k, v) })
	}
	kit.For(r.Cookies(), func(k, v string) { m.Optionv(k, v) })
	m.Options(ice.MSG_METHOD, r.Method, ice.MSG_COUNT, "0")
	m.Options(ice.MSG_USERWEB, _serve_domain(m), ice.MSG_USERPOD, m.Option(ice.POD))
	m.Options(ice.MSG_USERUA, r.Header.Get(UserAgent), ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	m.Options(ice.MSG_SESSID, kit.Select(m.Option(ice.MSG_SESSID), m.Option(CookieName(m.Option(ice.MSG_USERWEB)))))
	kit.If(m.Optionv(ice.MSG_CMDS) == nil, func() {
		kit.If(strings.TrimPrefix(r.URL.Path, key), func(p string) { m.Optionv(ice.MSG_CMDS, strings.Split(p, nfs.PS)) })
	})
	m.W.Header().Add(strings.ReplaceAll(ice.LOG_TRACEID, ".", "-"), m.Option(ice.LOG_TRACEID))
	defer func() { Render(m, m.Option(ice.MSG_OUTPUT), kit.List(m.Optionv(ice.MSG_ARGS))...) }()
	if cmds, ok := _serve_auth(m, key, kit.Simple(m.Optionv(ice.MSG_CMDS)), w, r); ok {
		defer func() {
			kit.If(m.Option(ice.MSG_STATUS) == "", func() { m.StatusTimeCount() })
			m.Cost(kit.Format("%s: %s %v", r.Method, r.URL.String(), m.FormatSize()))
		}()
		m.Option(ice.MSG_OPTS, kit.Simple(m.Optionv(ice.MSG_OPTION), func(k string) bool { return !strings.HasPrefix(k, ice.MSG_SESSID) }))
		if m.Detailv(m.PrefixKey(), cmds); len(cmds) > 1 && cmds[0] == ctx.ACTION {
			if !kit.IsIn(cmds[1], aaa.LOGIN, ctx.RUN, ctx.COMMAND) && m.Warn(r.Method == http.MethodGet, ice.ErrNotAllow) {
				return
			}
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
	if aaa.SessCheck(m, m.Option(ice.MSG_SESSID)); m.Option(ice.MSG_USERNAME) == "" {
		ls := kit.Simple(mdb.Cache(m, m.Option(ice.MSG_USERIP), func() ice.Any {
			if IsLocalHost(m) {
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
	return cmds, aaa.Right(m, key, cmds)
}

const (
	SSO    = "sso"
	URL    = "url"
	HTTP   = "http"
	HTTPS  = "https"
	DOMAIN = "domain"
	ORIGIN = "origin"
	FORM   = "form"
	BODY   = "body"

	SERVE_START = "serve.start"
)
const SERVE = "serve"

func init() {
	Index.MergeCommands(ice.Commands{P(ice.EXIT): {Hand: func(m *ice.Message, arg ...string) { m.Cmd(ice.EXIT) }},
		SERVE: {Name: "serve name auto main host dark system", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { cli.NodeInfo(m, ice.Info.Pathname, WORKER) }},
			ice.MAIN: {Name: "main index", Help: "首页", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ctx.INDEX) == "" {
					mdb.Config(m, ice.MAIN, "")
				} else {
					mdb.Config(m, ice.MAIN, CHAT_CMD+m.Option(ctx.INDEX)+nfs.PS)
				}
			}},
			log.TRACEID: {Help: "日志", Hand: func(m *ice.Message, arg ...string) {
				kit.If(len(arg) > 0, func() { ice.Info.Traceid = arg[0] })
				m.Echo(ice.Info.Traceid)
			}},
			tcp.HOST: {Help: "公网", Hand: func(m *ice.Message, arg ...string) { m.Echo(kit.Formats(PublicIP(m))) }},
			cli.DARK: {Help: "主题", Hand: func(m *ice.Message, arg ...string) {
				kit.If(tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)), func() {
					m.Cmd(cli.SYSTEM, "osascript", "-e", `tell app "System Events" to tell appearance preferences to set dark mode to not dark mode`)
				})
			}},
			cli.SYSTEM: {Help: "系统", Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, "System Settings.app") }},
			cli.START:  {Name: "start dev proto host port=9020 nodename username usernick", Hand: func(m *ice.Message, arg ...string) { _serve_start(m) }},
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(ice.DEMO) == ice.TRUE, func() { m.Cmd(CHAT_HEADER, ice.DEMO) })
				m.Go(func() {
					ssh.PrintQRCode(m, tcp.PublishLocalhost(m, _serve_address(m)))
					cli.Opens(m, mdb.Config(m, cli.OPEN))
				})
			}},
		}, gdb.EventsAction(SERVE_START), mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,status,name,proto,host,port"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Options(ice.MSG_ACTION, "").StatusTimeCount(kit.Dict(ice.MAIN, mdb.Config(m, ice.MAIN)))
		}},
	})
	ice.AddMergeAction(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) {
		if strings.HasPrefix(sub, nfs.PS) {
			actions := ice.Actions{}
			for k, v := range cmd.Actions {
				kit.If(!kit.IsIn(k, ice.CTX_INIT, ice.CTX_EXIT), func() { actions[k] = v })
			}
			kit.If(action.Hand == nil, func() { action.Hand = cmd.Hand })
			sub = kit.Select(P(key, sub), PP(key, sub), strings.HasSuffix(sub, nfs.PS))
			c.Commands[sub] = &ice.Command{Name: kit.Select(cmd.Name, action.Name), Actions: actions, Hand: func(m *ice.Message, arg ...string) {
				msg := m.Spawn(c, key, cmd)
				action.Hand(msg, arg...)
				m.Copy(msg)
			}, RawHand: action.Hand}
		}
	})
}
func Domain(host, port string) string { return kit.Format("%s://%s:%s", HTTP, host, port) }
func Script(m *ice.Message, str string, arg ...ice.Any) string {
	return ice.Render(m, ice.RENDER_SCRIPT, kit.Format(str, arg...))
}
func ChatCmdPath(m *ice.Message, arg ...string) string {
	if p := m.Option(ice.MSG_USERPOD); p != "" {
		return path.Join(CHAT_POD, p, "/cmd/", kit.Select(m.PrefixKey(), path.Join(arg...)))
	}
	return path.Join(CHAT_CMD, kit.Select(m.PrefixKey(), path.Join(arg...)))
}
func RequireFile(m *ice.Message, file string) string {
	if strings.HasPrefix(file, nfs.PS) || strings.HasPrefix(file, ice.HTTP) {
		return file
	} else if file != "" {
		return nfs.REQUIRE + file
	}
	return ""
}
func IsLocalHost(m *ice.Message) bool {
	return (m.R == nil || m.R.Header.Get("X-Forwarded-For") == "") && tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP))
}
