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
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _serve_start(m *ice.Message) {
	if m.Option(aaa.USERNAME) != "" {
		aaa.UserRoot(m, m.Option(aaa.USERNAME), m.Option(aaa.USERNICK))
	}
	if cli.NodeInfo(m, kit.Select(ice.Info.Hostname, m.Option("nodename")), SERVER); m.Option(tcp.PORT) == tcp.RANDOM {
		m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
	}
	m.Target().Start(m, m.OptionSimple(tcp.HOST, tcp.PORT)...)
	m.Go(func() { m.Sleep("1s").Cmd(BROAD, SERVE, m.OptionSimple(tcp.PORT)) })
	for _, v := range kit.Split(m.Option(ice.DEV)) {
		m.Cmd(SPACE, tcp.DIAL, ice.DEV, v, mdb.NAME, ice.Info.NodeName)
	}
}
func _serve_main(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	const (
		X_REAL_IP       = "X-Real-Ip"
		X_REAL_PORT     = "X-Real-Port"
		X_FORWARDED_FOR = "X-Forwarded-For"
		INDEX_MODULE    = "Index-Module"
		MOZILLA         = "Mozilla/5.0"
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
	if m.Logs(r.Header.Get(ice.MSG_USERIP), r.Method, r.URL.String()); m.Config(LOGHEADERS) == ice.TRUE {
		kit.Fetch(r.Header, func(k string, v []string) { m.Logs("Header", k, v) })
	}
	if r.Method == http.MethodGet && r.URL.Path != PP(SPACE) && !strings.HasPrefix(r.URL.Path, "/code/bash") {
		repos := kit.Select(ice.INTSHELL, ice.VOLCANOS, strings.Contains(r.Header.Get(UserAgent), MOZILLA))
		if p := path.Join(ice.USR, repos, r.URL.Path); r.URL.Path != ice.PS && nfs.ExistsFile(m, p) {
			Render(m.Spawn(w, r), ice.RENDER_DOWNLOAD, p)
			return false
		} else if msg := gdb.Event(m.Spawn(w, r), SERVE_REWRITE, r.Method, r.URL.Path, path.Join(m.Conf(SERVE, kit.Keym(repos, nfs.PATH)), r.URL.Path), repos); msg.Option(ice.MSG_OUTPUT) != "" {
			Render(msg, msg.Option(ice.MSG_OUTPUT), kit.List(msg.Optionv(ice.MSG_ARGS))...)
			return false
		}
	}
	return true
}
func _serve_handle(key string, cmd *ice.Command, m *ice.Message, w http.ResponseWriter, r *http.Request) {
	if u, e := url.Parse(r.Header.Get(Referer)); e == nil && r.URL.Path != PP(SPACE) {
		add := func(k, v string) { m.Logs("path", k, m.Option(k, v)) }
		switch arg := strings.Split(strings.TrimPrefix(u.Path, ice.PS), ice.PS); arg[0] {
		case "share":
			add(arg[0], arg[1])
		case "chat":
			for i := 1; i < len(arg)-1; i += 2 {
				add(arg[i], arg[i+1])
			}
		}
		// gdb.Event(m, SERVE_PARSE, strings.Split(strings.TrimPrefix(u.Path, ice.PS), ice.PS))
		kit.Fetch(u.Query(), func(k string, v []string) { m.Logs("Refer", k, v).Optionv(k, v) })
	}
	m.Option(ice.MSG_USERUA, r.Header.Get(UserAgent))
	for k, v := range kit.ParseQuery(r.URL.RawQuery) {
		kit.If(m.IsCliUA(), func() { v = kit.Simple(v, func(v string) (string, error) { return url.QueryUnescape(v) }) })
		m.Optionv(k, v)
	}
	switch r.Header.Get(ContentType) {
	case ContentJSON:
		data := kit.UnMarshal(r.Body)
		m.Logs(mdb.IMPORT, mdb.VALUE, kit.Format(data)).Optionv(ice.MSG_USERDATA, data)
		kit.Fetch(data, func(k string, v ice.Any) { m.Optionv(k, v) })
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(ContentLength))))
		kit.Fetch(r.PostForm, func(k string, v []string) {
			kit.If(m.IsCliUA(), func() { v = kit.Simple(v, func(v string) (string, error) { return url.QueryUnescape(v) }) })
			m.Logs("Form", k, kit.Join(v, ice.SP)).Optionv(k, v)
		})
	}
	kit.Fetch(r.Cookies(), func(k, v string) { m.Optionv(k, v) })
	m.OptionDefault(ice.MSG_HEIGHT, "480", ice.MSG_WIDTH, "320")
	m.Option(ice.MSG_USERUA, r.Header.Get(UserAgent))
	m.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	m.Option(ice.MSG_USERADDR, kit.Select(r.RemoteAddr, r.Header.Get(ice.MSG_USERADDR)))
	if m.Option(ice.MSG_USERWEB, _serve_domain(m)); m.Option(ice.POD) != "" {
		m.Option(ice.MSG_USERPOD, m.Option(ice.POD))
	}
	if u := OptionUserWeb(m); strings.Contains(u.Host, tcp.LOCALHOST) {
		m.Option(ice.MSG_USERHOST, tcp.PublishLocalhost(m, u.Scheme+"://"+u.Host))
	} else {
		m.Option(ice.MSG_USERHOST, u.Scheme+"://"+u.Host)
	}
	m.Option(ice.MSG_SESSID, kit.Select(m.Option(ice.MSG_SESSID), m.Option(CookieName(m.Option(ice.MSG_USERWEB)))))
	if m.Optionv(ice.MSG_CMDS) == nil {
		if p := strings.TrimPrefix(r.URL.Path, key); p != "" {
			m.Optionv(ice.MSG_CMDS, strings.Split(p, ice.PS))
		}
	}
	if cmds, ok := _serve_login(m, key, kit.Simple(m.Optionv(ice.MSG_CMDS)), w, r); ok {
		defer func() { m.Cost(kit.Format("%s %v %v", r.URL.Path, cmds, m.FormatSize())) }()
		m.Option(ice.MSG_OPTS, kit.Simple(m.Optionv(ice.MSG_OPTION), func(k string) bool { return !strings.HasPrefix(k, ice.MSG_SESSID) }))
		if m.Detailv(m.PrefixKey(), cmds); len(cmds) > 1 && cmds[0] == ctx.ACTION {
			m.ActionHand(cmd, key, cmds[1], cmds[2:]...)
		} else {
			m.CmdHand(cmd, key, cmds...)
		}
	}
	Render(m, m.Option(ice.MSG_OUTPUT), m.Optionv(ice.MSG_ARGS))
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
func _serve_login(m *ice.Message, key string, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	if aaa.SessCheck(m, m.Option(ice.MSG_SESSID)); m.Option(ice.MSG_USERNAME) == "" && r.URL.Path != PP(SPACE) && !strings.HasPrefix(r.URL.Path, "/sync") {
		gdb.Event(m, SERVE_LOGIN)
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
	SERVE_START   = "serve.start"
	SERVE_REWRITE = "serve.rewrite"
	SERVE_PARSE   = "serve.parse"
	SERVE_LOGIN   = "serve.login"
	SERVE_CHECK   = "serve.check"
	SERVE_STOP    = "serve.stop"

	WEB_LOGIN = "_login"
	DOMAIN    = "domain"
	INDEX     = "index"
	SSO       = "sso"
)
const SERVE = "serve"

func init() {
	Index.MergeCommands(ice.Commands{
		SERVE: {Name: "serve name auto start", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { cli.NodeInfo(m, ice.Info.Pathname, WORKER) }},
			cli.START: {Name: "start dev name=web proto=http host port=9020 nodename username usernick", Hand: func(m *ice.Message, arg ...string) {
				_serve_start(m)
			}},
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					opened := false
					m.Sleep("1s").Cmd(SPACE, func(values ice.Maps) {
						if values[mdb.TYPE] == CHROME {
							opened = true
						}
					})
					if opened {
						return
					}
					switch host := "http://localhost:" + m.Option(tcp.PORT); runtime.GOOS {
					case cli.WINDOWS:
						m.Cmd(cli.SYSTEM, "explorer.exe", host)
					case cli.DARWIN:
						m.Cmd(cli.SYSTEM, "open", host)
					}
				})
			}},
			SERVE_REWRITE: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] != http.MethodGet {
					return
				}
				switch arg[1] {
				case ice.PS:
					if arg[3] == ice.INTSHELL {
						RenderIndex(m, arg[3])
					} else {
						RenderMain(m, "", "")
					}
				default:
					if nfs.ExistsFile(m, arg[2]) {
						m.RenderDownload(arg[2])
					}
				}
			}},
			SERVE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_USERNAME) == "" && m.Config(tcp.LOCALHOST) == ice.TRUE && tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
					aaa.UserRoot(m)
				}
			}},
			DOMAIN: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					m.Config(tcp.LOCALHOST, ice.FALSE)
					ice.Info.Domain = arg[0]
				}
				m.Echo(ice.Info.Domain)
			}},
		}, mdb.HashAction(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,status,name,proto,host,port", tcp.LOCALHOST, ice.TRUE, LOGHEADERS, ice.FALSE,
			ice.INTSHELL, kit.Dict(nfs.PATH, ice.USR_INTSHELL, INDEX, ice.INDEX_SH, nfs.REPOS, "https://shylinux.com/x/intshell", nfs.BRANCH, nfs.MASTER),
			ice.VOLCANOS, kit.Dict(nfs.PATH, ice.USR_VOLCANOS, INDEX, "page/index.html", nfs.REPOS, "https://shylinux.com/x/volcanos", nfs.BRANCH, nfs.MASTER),
		), mdb.ClearHashOnExitAction(), ServeAction())},
		PP(ice.INTSHELL): {Name: "/intshell/", Help: "命令行", Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			RenderIndex(m, ice.INTSHELL, arg...)
		}},
		PP(ice.VOLCANOS): {Name: "/volcanos/", Help: "浏览器", Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			RenderIndex(m, ice.VOLCANOS, arg...)
		}},
		PP(ice.PUBLISH): {Name: "/publish/", Help: "定制化", Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			_share_local(m, ice.USR_PUBLISH, path.Join(arg...))
		}},
		PP(ice.REQUIRE): {Name: "/require/shylinux.com/x/volcanos/proto.js", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			cache := kit.Select(ice.USR_REQUIRE, m.Cmdx(cli.SYSTEM, "go", "env", "GOMODCACHE"))
			p := path.Join(cache, path.Join(arg...))
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
		PP(ice.REQUIRE, ice.NODE_MODULES): {Name: "/require/node_modules/", Help: "依赖库", Hand: func(m *ice.Message, arg ...string) {
			p := path.Join(ice.USR, ice.NODE_MODULES, path.Join(arg...))
			if !nfs.ExistsFile(m, p) {
				m.Cmd(cli.SYSTEM, "npm", "install", arg[0], kit.Dict(cli.CMD_DIR, ice.USR))
			}
			m.RenderDownload(p)
		}},
		PP(ice.REQUIRE, ice.USR): {Name: "/require/usr/", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			_share_local(m, ice.USR, path.Join(arg...))
		}},
		PP(ice.REQUIRE, ice.SRC): {Name: "/require/src/", Help: "源代码", Hand: func(m *ice.Message, arg ...string) {
			_share_local(m, ice.SRC, path.Join(arg...))
		}},
		PP(ice.HELP): {Name: "/help/", Help: "帮助", Actions: ice.MergeActions(ctx.CmdAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "tutor.shy")
			}
			if len(arg) > 0 && arg[0] != ctx.ACTION {
				arg[0] = path.Join(ice.SRC_HELP, arg[0])
			}
			m.Cmdy("web.chat./cmd/", arg)
		}},
	})
	ice.AddMerges(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) (ice.Handler, ice.Handler) {
		if strings.HasPrefix(sub, ice.PS) {
			if sub = kit.Select(sub, PP(key), sub == ice.PS); action.Hand == nil {
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
		return nil, nil
	})
}
func ServeAction() ice.Actions {
	return gdb.EventsAction(SERVE_START, SERVE_REWRITE, SERVE_PARSE, SERVE_LOGIN, SERVE_CHECK, SERVE_STOP)
}
