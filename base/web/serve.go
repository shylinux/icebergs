package web

import (
	"net/http"
	"net/url"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _serve_address(m *ice.Message) string {
	return kit.Format("http://localhost:%s", m.Option(tcp.PORT))
}
func _serve_start(m *ice.Message) {
	defer kit.For(kit.Split(m.Option(ice.DEV)), func(v string) { m.Sleep("10ms").Cmd(SPACE, tcp.DIAL, ice.DEV, v, mdb.NAME, ice.Info.NodeName) })
	kit.If(m.Option(aaa.USERNAME), func() { aaa.UserRoot(m, m.Option(aaa.USERNICK), m.Option(aaa.USERNAME)) })
	kit.If(m.Option(tcp.PORT) == tcp.RANDOM, func() { m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT)) })
	kit.If(runtime.GOOS == cli.WINDOWS, func() { m.Cmd(SPIDE, ice.OPS, _serve_address(m)+"/exit").Sleep300ms() })
	cli.NodeInfo(m, kit.Select(ice.Info.Hostname, m.Option(tcp.NODENAME)), SERVER)
	m.Target().Start(m, m.OptionSimple(tcp.HOST, tcp.PORT)...)
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
			r.Header.Set(ice.MSG_USERADDR, ip+ice.DF+r.Header.Get(X_REAL_PORT))
		}
	} else if ip := r.Header.Get(X_FORWARDED_FOR); ip != "" {
		r.Header.Set(ice.MSG_USERIP, kit.Split(ip)[0])
	} else if strings.HasPrefix(r.RemoteAddr, "[") {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, "]")[0][1:])
	} else {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, ice.DF)[0])
	}
	if m.Logs(r.Header.Get(ice.MSG_USERIP), r.Method, r.URL.String()); r.Method == http.MethodGet {
		if msg := m.Spawn(w, r).Options(ice.MSG_USERUA, r.UserAgent()); path.Join(r.URL.Path) == ice.PS {
			return !Render(RenderMain(msg), msg.Option(ice.MSG_OUTPUT), kit.List(msg.Optionv(ice.MSG_ARGS))...)
		} else if p := path.Join(kit.Select(ice.USR_VOLCANOS, ice.USR_INTSHELL, msg.IsCliUA()), r.URL.Path); nfs.ExistsFile(msg, p) {
			return !Render(msg, ice.RENDER_DOWNLOAD, p)
		}
	}
	return true
}
func _serve_handle(key string, cmd *ice.Command, m *ice.Message, w http.ResponseWriter, r *http.Request) {
	_log := func(level string, arg ...ice.Any) *ice.Message { return m.Logs(strings.Title(level), arg...) }
	if u, e := url.Parse(r.Header.Get(Referer)); e == nil {
		add := func(k, v string) { _log(nfs.PATH, k, m.Option(k, v)) }
		switch arg := strings.Split(strings.TrimPrefix(u.Path, ice.PS), ice.PS); arg[0] {
		case CHAT:
			kit.For(arg[1:], func(k, v string) { add(k, v) })
		case SHARE:
			add(arg[0], arg[1])
		}
		kit.For(u.Query(), func(k string, v []string) { _log(ctx.ARGS, k, v).Optionv(k, v) })
	}
	m.Options(ice.MSG_USERUA, r.Header.Get(UserAgent))
	for k, v := range kit.ParseQuery(r.URL.RawQuery) {
		kit.If(m.IsCliUA(), func() { v = kit.Simple(v, func(v string) (string, error) { return url.QueryUnescape(v) }) })
		m.Optionv(k, v)
	}
	switch r.Header.Get(ContentType) {
	case ContentJSON:
		data := kit.UnMarshal(r.Body)
		_log("Body", mdb.VALUE, kit.Format(data)).Optionv(ice.MSG_USERDATA, data)
		kit.For(data, func(k string, v ice.Any) { m.Optionv(k, v) })
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(ContentLength))))
		kit.For(r.PostForm, func(k string, v []string) {
			kit.If(m.IsCliUA(), func() { v = kit.Simple(v, func(v string) (string, error) { return url.QueryUnescape(v) }) })
			_log("Form", k, kit.Join(v, ice.SP)).Optionv(k, v)
		})
	}
	kit.For(r.Cookies(), func(k, v string) { m.Optionv(k, v) })
	m.OptionDefault(ice.MSG_HEIGHT, "480", ice.MSG_WIDTH, "320")
	m.Options(ice.MSG_USERWEB, _serve_domain(m), ice.MSG_USERPOD, m.Option(ice.POD))
	m.Options(ice.MSG_SESSID, kit.Select(m.Option(ice.MSG_SESSID), m.Option(CookieName(m.Option(ice.MSG_USERWEB)))))
	m.Options(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP), ice.MSG_USERADDR, kit.Select(r.RemoteAddr, r.Header.Get(ice.MSG_USERADDR)))
	if m.Optionv(ice.MSG_CMDS) == nil {
		if p := strings.TrimPrefix(r.URL.Path, key); p != "" {
			m.Optionv(ice.MSG_CMDS, strings.Split(p, ice.PS))
		}
	}
	if cmds, ok := _serve_login(m, key, kit.Simple(m.Optionv(ice.MSG_CMDS)), w, r); ok {
		defer func() { m.Cost(kit.Format("%s: %s %v", r.Method, m.PrefixPath()+path.Join(cmds...), m.FormatSize())) }()
		m.Option(ice.MSG_OPTS, kit.Simple(m.Optionv(ice.MSG_OPTION), func(k string) bool { return !strings.HasPrefix(k, ice.MSG_SESSID) }))
		if m.Detailv(m.PrefixKey(), cmds); len(cmds) > 1 && cmds[0] == ctx.ACTION {
			m.ActionHand(cmd, key, cmds[1], cmds[2:]...)
		} else {
			m.CmdHand(cmd, key, cmds...)
		}
	}
	Render(m, m.Option(ice.MSG_OUTPUT), kit.List(m.Optionv(ice.MSG_ARGS))...)
}
func _serve_domain(m *ice.Message) string {
	return kit.GetValid(
		func() string { return kit.Select("", m.R.Header.Get(Referer), m.R.Method == http.MethodPost) },
		func() string { return m.R.Header.Get("X-Host") },
		func() string { return ice.Info.Domain },
		func() string {
			if b, e := regexp.MatchString("^[0-9.]+$", m.R.Host); b && e == nil {
				return kit.Format("%s://%s:%s", kit.Select("https", ice.HTTP, m.R.TLS == nil), m.R.Host, m.Option(tcp.PORT))
			}
			return kit.Format("%s://%s", kit.Select("https", ice.HTTP, m.R.TLS == nil), m.R.Host)
		},
	)
}

var localhost = sync.Map{}

func _serve_login(m *ice.Message, key string, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	if r.URL.Path == PP(SPACE) {
		return cmds, true
	}
	if aaa.SessCheck(m, m.Option(ice.MSG_SESSID)); m.Option(ice.MSG_USERNAME) == "" {
		last, ok := localhost.Load(m.Option(ice.MSG_USERIP))
		if ls := kit.Simple(last); ok && len(ls) > 0 && kit.Time(m.Time())-kit.Time(ls[0]) < int64(time.Hour) {
			m.Auth(
				aaa.USERNICK, m.Option(ice.MSG_USERNICK, ls[1]),
				aaa.USERNAME, m.Option(ice.MSG_USERNAME, ls[2]),
				aaa.USERROLE, m.Option(ice.MSG_USERROLE, ls[3]),
				"last", ls[0],
			)
		} else if ice.Info.Localhost && tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
			aaa.UserRoot(m)
		} else {
			gdb.Event(m, SERVE_LOGIN)
		}
		if ice.Info.Localhost {
			localhost.Store(m.Option(ice.MSG_USERIP), kit.Simple(m.Time(), m.OptionSplit(ice.MSG_USERNICK, ice.MSG_USERNAME, ice.MSG_USERROLE)))
		}
	}
	if _, ok := m.Target().Commands[WEB_LOGIN]; ok {
		return cmds, !m.Target().Cmd(m, WEB_LOGIN, kit.Simple(key, cmds)...).IsErr()
	} else if gdb.Event(m, SERVE_CHECK, key, cmds); m.IsErr() {
		return cmds, false
	} else if m.IsOk() {
		return cmds, m.SetResult() != nil
	} else {
		return cmds, aaa.Right(m, key, cmds)
	}
}

const (
	SERVE_START = "serve.start"
	SERVE_LOGIN = "serve.login"
	SERVE_CHECK = "serve.check"

	REQUIRE_SRC     = "require/src/"
	REQUIRE_USR     = "require/usr/"
	REQUIRE_MODULES = "require/modules/"

	WEB_LOGIN = "_login"
	DOMAIN    = "domain"
	INDEX     = "index"
	SSO       = "sso"
)
const SERVE = "serve"

func init() {
	Index.MergeCommands(ice.Commands{"/exit": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(ice.EXIT) }},
		SERVE: {Name: "serve name auto start", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				ice.Info.Localhost = mdb.Config(m, tcp.LOCALHOST) == ice.TRUE
				cli.NodeInfo(m, ice.Info.Pathname, WORKER)
			}},
			cli.START: {Name: "start dev proto host port=9020 nodename username usernick", Hand: func(m *ice.Message, arg ...string) {
				_serve_start(m)
			}},
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					msg := m.Spawn(kit.Dict(ice.LOG_DISABLE, ice.TRUE)).Sleep("10ms")
					msg.Cmd(ssh.PRINTF, kit.Dict(nfs.CONTENT, ice.NL+ice.Render(msg, ice.RENDER_QRCODE, tcp.PublishLocalhost(msg, kit.Format("http://localhost:%s", msg.Option(tcp.PORT))))))
					msg.Cmd(ssh.PROMPT)
					opened := false
					for i := 0; i < 3 && !opened; i++ {
						msg.Sleep("1s").Cmd(SPACE, func(values ice.Maps) { kit.If(values[mdb.TYPE] == CHROME, func() { opened = true }) })
					}
					kit.If(!opened, func() { cli.Opens(msg, _serve_address(msg)) })
				})
			}},
			DOMAIN: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(len(arg) > 0, func() { ice.Info.Domain, ice.Info.Localhost = arg[0], false })
				m.Echo(ice.Info.Domain)
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,status,name,proto,host,port", tcp.LOCALHOST, ice.TRUE), mdb.ClearOnExitHashAction())},
		PP(ice.PUBLISH): {Name: "/publish/", Help: "定制化", Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			_share_local(m, ice.USR_PUBLISH, path.Join(arg...))
		}},
		PP(ice.REQUIRE): {Name: "/require/shylinux.com/x/volcanos/proto.js", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 4 {
				m.RenderStatusBadRequest()
				return
			} else if path.Join(arg[:3]...) == ice.Info.Make.Module && nfs.ExistsFile(m, path.Join(arg[3:]...)) {
				m.RenderDownload(path.Join(arg[3:]...))
				return
			}
			p := path.Join(kit.Select(ice.USR_REQUIRE, m.Cmdx(cli.SYSTEM, "go", "env", "GOMODCACHE")), path.Join(arg...))
			if !nfs.ExistsFile(m, p) {
				if p = path.Join(ice.USR_REQUIRE, path.Join(arg...)); !nfs.ExistsFile(m, p) {
					ls := strings.SplitN(path.Join(arg[:3]...), ice.AT, 2)
					if v := kit.Select(ice.Info.Gomod[ls[0]], ls, 1); v == "" {
						m.Cmd(cli.SYSTEM, "git", "clone", "https://"+ls[0], path.Join(ice.USR_REQUIRE, path.Join(arg[:3]...)))
					} else {
						m.Cmd(cli.SYSTEM, "git", "clone", "-b", v, "https://"+ls[0], path.Join(ice.USR_REQUIRE, path.Join(arg[:3]...)))
					}
				}
			}
			m.RenderDownload(p)
		}},
		PP(REQUIRE_SRC): {Name: "/require/src/", Help: "源代码", Actions: ice.MergeActions(ctx.CmdAction(), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			_share_local(m, ice.SRC, path.Join(arg...))
		}},
		PP(REQUIRE_USR): {Name: "/require/usr/", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			_share_local(m, ice.USR, path.Join(arg...))
		}},
		PP(REQUIRE_MODULES): {Name: "/require/modules/", Help: "依赖库", Hand: func(m *ice.Message, arg ...string) {
			p := path.Join(ice.USR_MODULES, path.Join(arg...))
			if !nfs.ExistsFile(m, p) {
				m.Cmd(cli.SYSTEM, "npm", "install", arg[0], kit.Dict(cli.CMD_DIR, ice.USR))
			}
			m.RenderDownload(p)
		}},
	})
	ice.AddMerges(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) {
		if strings.HasPrefix(sub, ice.PS) {
			if sub = kit.Select(PP(key, sub), PP(key), sub == ice.PS); action.Hand == nil {
				action.Hand = func(m *ice.Message, arg ...string) { m.Cmdy(key, arg) }
			}
			actions := ice.Actions{}
			for k, v := range cmd.Actions {
				switch k {
				case ctx.COMMAND, ice.RUN:
					actions[k] = v
				}
			}
			c.Commands[sub] = &ice.Command{Name: sub, Help: cmd.Help, Actions: actions, Hand: action.Hand}
		}
	})
}
func ServeAction() ice.Actions { return gdb.EventsAction(SERVE_START, SERVE_LOGIN, SERVE_CHECK) }
