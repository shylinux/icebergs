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
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _serve_start(m *ice.Message) {
	if cli.NodeInfo(m, kit.Select(ice.Info.HostName, m.Option("nodename")), SERVER); m.Option(tcp.PORT) == tcp.RANDOM {
		m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
	}
	aaa.UserRoot(m, m.Option(aaa.USERNAME), m.Option(aaa.USERNICK))
	m.Target().Start(m, m.OptionSimple(tcp.HOST, tcp.PORT)...)
	m.Sleep300ms().Go(func() { m.Cmd(BROAD, SERVE) })
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
	if m.Config(LOGHEADERS) == ice.TRUE {
		for k, v := range r.Header {
			m.Info("%s: %v", k, kit.Format(v), meta)
		}
		m.Info("", meta)
	}
	repos := kit.Select(ice.INTSHELL, ice.VOLCANOS, strings.Contains(r.Header.Get(UserAgent), "Mozilla/5.0"))
	if msg := gdb.Event(m.Spawn(w, r), SERVE_REWRITE, r.Method, r.URL.Path, path.Join(m.Conf(SERVE, kit.Keym(repos, nfs.PATH)), r.URL.Path), repos); msg.Option(ice.MSG_OUTPUT) != "" {
		Render(msg, msg.Option(ice.MSG_OUTPUT), kit.List(msg.Optionv(ice.MSG_ARGS))...)
		return false
	}
	return true
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
func _serve_handle(key string, cmd *ice.Command, m *ice.Message, w http.ResponseWriter, r *http.Request) {
	meta := logs.FileLineMeta("")
	if u, e := url.Parse(r.Header.Get(Referer)); e == nil {
		gdb.Event(m, SERVE_PARSE, strings.Split(strings.TrimPrefix(u.Path, ice.PS), ice.PS))
		kit.Fetch(u.Query(), func(k string, v []string) { m.Logs("refer", k, v, meta).Optionv(k, v) })
	}
	switch r.Header.Get(ContentType) {
	case ContentJSON:
		defer r.Body.Close()
		var data ice.Any
		if e := json.NewDecoder(r.Body).Decode(&data); !m.Warn(e, ice.ErrNotFound, data) {
			m.Logs(mdb.IMPORT, mdb.VALUE, kit.Format(data))
			m.Optionv(ice.MSG_USERDATA, data)
		}
		kit.Fetch(data, func(key string, value ice.Any) { m.Optionv(key, value) })
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(ContentLength))))
		if r.ParseForm(); len(r.PostForm) > 0 {
			kit.Fetch(r.PostForm, func(k string, v []string) {
				if len(v) > 1 {
					m.Logs("form", k, len(v), kit.Join(v, ice.SP), meta)
				} else {
					m.Logs("form", k, v, meta)
				}
			})
		}
	}
	m.R, m.W = r, w
	for k, v := range r.Form {
		if m.IsCliUA() {
			for i, p := range v {
				v[i], _ = url.QueryUnescape(p)
			}
		}
		m.Optionv(k, v)
	}
	for k, v := range r.PostForm {
		m.Optionv(k, v)
	}
	for _, v := range r.Cookies() {
		m.Optionv(v.Name, v.Value)
	}
	m.Option(ice.MSG_USERADDR, kit.Select(r.RemoteAddr, r.Header.Get(ice.MSG_USERADDR)))
	m.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	m.Option(ice.MSG_USERUA, r.Header.Get(UserAgent))
	if m.Option(ice.MSG_USERWEB, _serve_domain(m)); m.Option(ice.POD) != "" {
		m.Option(ice.MSG_USERPOD, m.Option(ice.POD))
	}
	if sessid := m.Option(CookieName(m.Option(ice.MSG_USERWEB))); m.Option(ice.MSG_SESSID) == "" {
		m.Option(ice.MSG_SESSID, sessid)
	}
	if m.Optionv(ice.MSG_CMDS) == nil {
		if p := strings.TrimPrefix(r.URL.Path, key); p != "" {
			m.Optionv(ice.MSG_CMDS, strings.Split(p, ice.PS))
		}
	}
	if cmds, ok := _serve_login(m, key, kit.Simple(m.Optionv(ice.MSG_CMDS)), w, r); ok {
		defer func() { m.Cost(kit.Format("%s %v %v", r.URL.Path, cmds, m.FormatSize())) }()
		m.Option(ice.MSG_OPTS, kit.Simple(m.Optionv(ice.MSG_OPTION), func(k string) bool { return !strings.HasPrefix(k, ice.MSG_SESSID) }))
		if len(cmds) > 0 && cmds[0] == ctx.ACTION {
			m.Target().Cmd(m, key, cmds...)
		} else {
			m.CmdHand(cmd, key, cmds...)
		}
	}
	gdb.Event(m, SERVE_RENDER, m.Option(ice.MSG_OUTPUT))
	Render(m, m.Option(ice.MSG_OUTPUT), m.Optionv(ice.MSG_ARGS))
}
func _serve_login(m *ice.Message, key string, cmds []string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	if aaa.SessCheck(m, m.Option(ice.MSG_SESSID)); m.Option(ice.MSG_USERNAME) == "" {
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
	SERVE_RENDER  = "serve.render"
	SERVE_STOP    = "serve.stop"

	WEB_LOGIN = "_login"
	SSO       = "sso"

	DOMAIN = "domain"
	INDEX  = "index"
)
const SERVE = "serve"

func init() {
	Index.MergeCommands(ice.Commands{
		SERVE: {Name: "serve name auto start spide", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				cli.NodeInfo(m, ice.Info.PathName, WORKER)
				aaa.White(m, LOGIN)
			}},
			cli.START: {Name: "start dev proto=http host port=9020 nodename username usernick", Hand: func(m *ice.Message, arg ...string) {
				_serve_start(m)
			}},
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					m.Sleep("30ms", ssh.PRINTF, kit.Dict(nfs.CONTENT, "\r"+ice.Render(m, ice.RENDER_QRCODE, m.Cmdx(SPACE, DOMAIN))+ice.NL))
					m.Cmd(ssh.PROMPT)
				})
			}},
			SERVE_REWRITE: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] != SPIDE_GET {
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
			SERVE_PARSE: {Hand: func(m *ice.Message, arg ...string) {
				m.Options(ice.HEIGHT, "480", ice.WIDTH, "320")
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
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,status,name,proto,host,port,dev", tcp.LOCALHOST, ice.TRUE, LOGHEADERS, ice.FALSE,
			ice.INTSHELL, kit.Dict(nfs.PATH, ice.USR_INTSHELL, INDEX, ice.INDEX_SH, nfs.REPOS, "https://shylinux.com/x/intshell", nfs.BRANCH, nfs.MASTER),
			ice.VOLCANOS, kit.Dict(nfs.PATH, ice.USR_VOLCANOS, INDEX, "page/index.html", nfs.REPOS, "https://shylinux.com/x/volcanos", nfs.BRANCH, nfs.MASTER),
		), ServeAction())},
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
			_share_local(m, ice.USR, path.Join(arg...))
		}},
		PP(ice.REQUIRE, ice.SRC): {Name: "/require/src/", Help: "源代码", Hand: func(m *ice.Message, arg ...string) {
			_share_local(m, ice.SRC, path.Join(arg...))
		}},
		PP(ice.HELP): {Name: "/help/", Help: "帮助", Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
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
			c.Commands[sub] = &ice.Command{Name: sub, Help: cmd.Help, Hand: action.Hand}
		}
		return nil, nil
	})
}
func ServeAction() ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		for sub := range m.Target().Commands[m.CommandKey()].Actions {
			if serveActions[sub] == ice.TRUE {
				gdb.Watch(m, sub)
			}
		}
	}}}
}

var serveActions = kit.DictList(SERVE_START, SERVE_REWRITE, SERVE_PARSE, SERVE_LOGIN, SERVE_CHECK, SERVE_RENDER, SERVE_STOP)
