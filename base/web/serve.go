package web

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _serve_address(m *ice.Message) string { return HostPort(m, tcp.LOCALHOST, m.Option(tcp.PORT)) }
func _serve_start(m *ice.Message) {
	kit.If(m.Option(aaa.USERNAME), func() { aaa.UserRoot(m, "", m.Option(aaa.USERNAME), m.Option(aaa.USERNICK), m.Option(aaa.LANGUAGE)) })
	kit.If(m.Option(tcp.PORT) == tcp.RANDOM, func() { m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT)) })
	m.Go(func() {
		m.Cmd(SPIDE, ice.OPS, _serve_address(m)+nfs.PS+ice.EXIT, ice.Maps{CLIENT_TIMEOUT: cli.TIME_30ms, ice.LOG_DISABLE: ice.TRUE})
	}).Sleep(cli.TIME_1s)
	cli.NodeInfo(m, kit.Select(kit.Split(ice.Info.Hostname, nfs.PT)[0], m.Option(tcp.NODENAME)), SERVER, mdb.Config(m, mdb.ICONS))
	kit.If(ice.HasVar(), func() { m.Cmd(nfs.SAVE, ice.VAR_LOG_ICE_PORT, m.Option(tcp.PORT)) })
	m.Spawn(ice.Maps{TOKEN: ""}).Start("", m.OptionSimple(tcp.HOST, tcp.PORT)...)
	if m.Cmd(tcp.HOST).Length() == 0 {
		return
	}
	kit.For(kit.Split(m.Option(ice.DEV)), func(dev string) {
		if strings.HasPrefix(dev, HTTP) {
			m.Cmd(SPIDE, mdb.CREATE, dev, ice.DEV, "", nfs.REPOS)
			m.Cmd(SPIDE, mdb.CREATE, dev, "dev_ip", "", "dev_ip")
			dev = ice.DEV
		}
		if msg := m.Cmd(SPIDE, dev); msg.Append(TOKEN) == "" {
			if m.Option(TOKEN) != "" {
				m.Sleep300ms(SPACE, tcp.DIAL, ice.DEV, dev, TOKEN, m.Option(TOKEN))
			} else {
				m.Sleep300ms(SPACE, tcp.DIAL, ice.DEV, dev)
			}
		}
	})
}
func _serve_main(m *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	const (
		X_REAL_IP    = "X-Real-Ip"
		X_REAL_PORT  = "X-Real-Port"
		INDEX_MODULE = "Index-Module"
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
	} else if ip := r.Header.Get(html.XForwardedFor); ip != "" {
		r.Header.Set(ice.MSG_USERIP, kit.Split(ip)[0])
	} else if strings.HasPrefix(r.RemoteAddr, "[") {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, "]")[0][1:])
	} else {
		r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, nfs.DF)[0])
	}
	if !kit.HasPrefix(r.URL.String(), nfs.VOLCANOS, nfs.REQUIRE_MODULES, nfs.INTSHELL) {
		r.Header.Set(ice.LOG_TRACEID, log.Traceid(m))
		m.Logs(r.Header.Get(ice.MSG_USERIP), r.Method, r.URL.String(), logs.TraceidMeta(r.Header.Get(ice.LOG_TRACEID)))
	}
	if path.Join(r.URL.Path) == nfs.PS && strings.HasPrefix(r.UserAgent(), html.Mozilla) {
		r.URL.Path = kit.Select(nfs.PS, mdb.Config(m, ice.MAIN))
	}
	if r.Method == http.MethodGet {
		msg := m.Spawn(w, r).Options(ice.MSG_USERUA, r.UserAgent(), ice.LOG_TRACEID, r.Header.Get(ice.LOG_TRACEID), ParseLink(m, kit.Select(r.URL.String(), r.Referer())))
		if path.Join(r.URL.Path) == nfs.PS {
			msg.Options(ice.MSG_USERWEB, _serve_domain(msg))
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
	// _serve_params(msg, r.Header.Get(html.Referer))
	if strings.HasPrefix(r.URL.Path, "/.git/") {
		return false
	}
	_serve_params(msg, r.URL.String())
	ispod := msg.Option(ice.POD) != ""
	if strings.HasPrefix(r.URL.Path, nfs.V) {
		return Render(msg, ice.RENDER_DOWNLOAD, path.Join(ice.USR_VOLCANOS, strings.TrimPrefix(r.URL.Path, nfs.V)))
	} else if kit.HasPrefix(r.URL.Path, nfs.P) {
		if kit.Contains(r.URL.String(), "render=replace") {
			return false
		}
		p := path.Join(strings.TrimPrefix(r.URL.Path, nfs.P))
		if pp := path.Join(nfs.USR_LOCAL_WORK, msg.Option(ice.POD)); ispod && nfs.Exists(msg, pp) && !strings.HasPrefix(p, "require/") {
			if kit.HasPrefix(p, "var/", "usr/local/") {
				return false
			}
			if pp = path.Join(pp, p); nfs.Exists(msg, pp) {
				return Render(msg, ice.RENDER_DOWNLOAD, pp)
			} else if nfs.Exists(msg, p) {
				return Render(msg, ice.RENDER_DOWNLOAD, p)
			}
		}
		if kit.HasPrefix(p, ice.USR_ICEBERGS, ice.USR_ICONS) && nfs.Exists(msg, p) {
			return Render(msg, ice.RENDER_DOWNLOAD, p)
		}
		if !ispod {
			return (kit.HasPrefix(p, nfs.SRC) && nfs.Exists(msg, p)) && Render(msg, ice.RENDER_DOWNLOAD, p)
		}
	} else if kit.HasPrefix(r.URL.Path, nfs.M) {
		p := nfs.USR_MODULES + strings.TrimPrefix(r.URL.Path, nfs.M)
		return nfs.Exists(msg, p) && Render(msg, ice.RENDER_DOWNLOAD, p)
	} else if kit.HasPrefix(path.Base(r.URL.Path), "MP_verify_") {
		return Render(msg, ice.RENDER_DOWNLOAD, nfs.ETC+path.Base(r.URL.Path))
	} else if p := path.Join(kit.Select(ice.USR_VOLCANOS, ice.USR_INTSHELL, msg.IsCliUA()), r.URL.Path); nfs.Exists(msg, p) {
		return Render(msg, ice.RENDER_DOWNLOAD, p)
	} else if p = path.Join(nfs.USR, r.URL.Path); kit.HasPrefix(r.URL.Path, nfs.VOLCANOS, nfs.INTSHELL) && nfs.Exists(msg, p) {
		return Render(msg, ice.RENDER_DOWNLOAD, p)
	}
	return false
	p := ""
	if p = strings.TrimPrefix(r.URL.Path, nfs.REQUIRE); kit.HasPrefix(r.URL.Path, nfs.REQUIRE_SRC, nfs.REQUIRE+ice.USR_ICONS, nfs.REQUIRE+ice.USR_ICEBERGS) && nfs.Exists(msg, p) {
		return !ispod && Render(msg, ice.RENDER_DOWNLOAD, p)
	} else if p = path.Join(nfs.USR_MODULES, strings.TrimPrefix(r.URL.Path, nfs.REQUIRE_MODULES)); kit.HasPrefix(r.URL.Path, nfs.REQUIRE_MODULES) && nfs.Exists(msg, p) {
		return Render(msg, ice.RENDER_DOWNLOAD, p)
	} else {
		return false
	}
}
func _serve_params(m *ice.Message, p string) {
	if u, e := url.Parse(p); e == nil {
		switch arg := strings.Split(strings.TrimPrefix(u.Path, nfs.PS), nfs.PS); arg[0] {
		case CHAT:
			kit.For(arg[1:], func(k, v string) { m.Option(k, v) })
		case SHARE:
			m.Option(arg[0], arg[1])
		case "s":
			m.Option(ice.POD, kit.Select("", arg, 1))
		}
		kit.For(u.Query(), func(k string, v []string) { m.Optionv(k, v) })
	}
}
func _serve_handle(key string, cmd *ice.Command, m *ice.Message, w http.ResponseWriter, r *http.Request) {
	debug := strings.Contains(r.URL.String(), "debug=true") || strings.Contains(r.Header.Get(html.Referer), "debug=true")
	m.Options(ice.LOG_DEBUG, ice.FALSE, ice.LOG_TRACEID, r.Header.Get(ice.LOG_TRACEID))
	_log := func(level string, arg ...ice.Any) *ice.Message {
		if debug || arg[0] == ice.MSG_CMDS {
			return m.Logs(strings.Title(level), arg...)
		}
		return m
	}
	kit.If(r.Header.Get(html.Referer), func(p string) { _log("page", html.Referer, p) })
	_serve_params(m, r.Header.Get(html.Referer))
	_serve_params(m, r.URL.String())
	if r.Method == http.MethodGet && m.Option(ice.MSG_CMDS) != "" {
		_log(ctx.ARGS, ice.MSG_CMDS, m.Optionv(ice.MSG_CMDS))
	}
	switch kit.Select("", kit.Split(r.Header.Get(html.ContentType)), 0) {
	case html.ApplicationJSON:
		buf, _ := ioutil.ReadAll(r.Body)
		m.Option("request.data", string(buf))
		kit.For(kit.UnMarshal(string(buf)), func(k string, v ice.Any) { m.Optionv(k, v) })
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select("4096", r.Header.Get(html.ContentLength))))
		kit.For(r.PostForm, func(k string, v []string) { _log(FORM, k, kit.Join(v, lex.SP)).Optionv(k, v) })
	}
	kit.For(r.Cookies(), func(k, v string) { m.Optionv(k, v) })
	m.Options(ice.MSG_METHOD, r.Method, ice.MSG_COUNT, "0")
	m.Options(ice.MSG_REFERER, r.Header.Get(html.Referer))
	m.Options(ice.MSG_USERWEB, _serve_domain(m), ice.MSG_USERPOD, m.Option(ice.POD))
	m.Options(ice.MSG_USERUA, r.Header.Get(html.UserAgent), ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	m.Options(ice.MSG_SESSID, kit.Select(m.Option(ice.MSG_SESSID), m.Option(CookieName(m.Option(ice.MSG_USERWEB)))))
	kit.If(m.Optionv(ice.MSG_CMDS) == nil, func() {
		kit.If(strings.TrimPrefix(r.URL.Path, key), func(p string) { m.Optionv(ice.MSG_CMDS, strings.Split(p, nfs.PS)) })
	})
	UserHost(m)
	for k, v := range m.R.Header {
		// m.Info("what %v %v", k, v)
		kit.If(strings.HasPrefix(k, "Wechatpay"), func() { m.Option(k, v) })
	}
	m.W.Header().Add(strings.ReplaceAll(ice.LOG_TRACEID, ".", "-"), m.Option(ice.LOG_TRACEID))
	defer func() { Render(m, m.Option(ice.MSG_OUTPUT), kit.List(m.Optionv(ice.MSG_ARGS))...) }()
	if cmds, ok := _serve_auth(m, key, kit.Simple(m.Optionv(ice.MSG_CMDS)), w, r); ok {
		m.Option(ice.MSG_COST, "")
		defer func() {
			kit.If(m.Option(ice.MSG_STATUS) == "", func() { m.StatusTimeCount() })
			m.Cost(kit.Format("%s: %s %v", r.Method, r.URL.String(), m.FormatSize())).Options(ice.MSG_COST, m.FormatCost())
		}()
		m.Option(ice.MSG_OPTS, kit.Simple(m.Optionv(ice.MSG_OPTION), func(k string) bool { return !strings.HasPrefix(k, ice.MSG_SESSID) }))
		if m.Detailv(m.ShortKey(), cmds); len(cmds) > 1 && cmds[0] == ctx.ACTION && cmds[1] != ctx.ACTION {
			if !kit.IsIn(cmds[1], aaa.LOGIN, ctx.RUN, ctx.COMMAND) && m.WarnNotAllow(r.Method == http.MethodGet) {
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
		func() string { return kit.Select("", m.R.Header.Get(html.Referer), m.R.Method == http.MethodPost) },
		func() string { return m.R.Header.Get(html.XHost) },
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
	kit.If(len(cmds) > 0, func() { cmds = append(kit.Split(cmds[0], ","), cmds[1:]...) })
	kit.If(!aaa.IsTechOrRoot(m), func() { m.Option("user_uid", "") })
	if r.URL.Path == PP(SPACE) {
		aaa.SessCheck(m, m.Option(ice.MSG_SESSID))
		return cmds, true
	}
	defer func() { m.Options(ice.MSG_CMDS, "") }()
	if strings.Contains(m.Option(ice.MSG_SESSID), " ") {
		m.Cmdy(kit.Split(m.Option(ice.MSG_SESSID)))
	} else if aaa.SessCheck(m, m.Option(ice.MSG_SESSID)); m.Option(ice.MSG_USERNAME) == "" {
		if ls := kit.Simple(mdb.Cache(m, m.Option(ice.MSG_USERIP), func() ice.Any {
			if !IsLocalHost(m) {
				return nil
			}
			aaa.UserRoot(m)
			return kit.Simple(m.Time(), m.OptionSplit(ice.MSG_USERNICK, ice.MSG_USERNAME, ice.MSG_USERROLE))
		})); len(ls) > 0 {
			aaa.SessAuth(m, kit.Dict(aaa.USERNICK, ls[1], aaa.USERNAME, ls[2], aaa.USERROLE, ls[3]), CACHE, ls[0])
		}
	}
	Count(m, aaa.IP, m.Option(ice.MSG_USERIP), m.Option(ice.MSG_USERUA))
	return cmds, aaa.Right(m, key, cmds)
}

const (
	SSO    = "sso"
	URL    = "url"
	HTTP   = "http"
	HTTPS  = "https"
	DOMAIN = "domain"
	FORM   = "form"
	BODY   = "body"
	HOME   = "home"

	SERVE_START = "serve.start"
	PROXY_CONF  = "proxyConf"
	PROXY_PATH  = "usr/local/daemon/10000/"
	PROXY_CMDS  = "./sbin/nginx"
)
const SERVE = "serve"

func init() {
	Index.MergeCommands(ice.Commands{P(ice.EXIT): {Hand: func(m *ice.Message, arg ...string) { m.Cmd(ice.EXIT) }},
		SERVE: {Name: "serve port auto main host system", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			ice.MAIN: {Name: "main index", Help: "首页", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ctx.INDEX) == "" {
					mdb.Config(m, ice.MAIN, "")
				} else {
					mdb.Config(m, ice.MAIN, C(m.Option(ctx.INDEX)))
				}
			}},
			mdb.ICONS:  {Hand: func(m *ice.Message, arg ...string) { mdb.Config(m, mdb.ICONS, arg[0]) }},
			tcp.HOST:   {Help: "公网", Hand: func(m *ice.Message, arg ...string) { m.Echo(kit.Formats(PublicIP(m))) }},
			cli.SYSTEM: {Help: "系统", Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, "System Settings.app") }},
			cli.START: {Name: "start dev proto host port=9020 nodename username usernick", Hand: func(m *ice.Message, arg ...string) {
				if runtime.GOOS == cli.LINUX {
					m.Cmd(nfs.SAVE, nfs.ETC_LOCAL_SH, m.Spawn(ice.Maps{cli.PWD: kit.Path(""), aaa.USER: kit.UserName(), ctx.ARGS: kit.JoinCmds(arg...)}).Template("local.sh")+lex.NL)
					m.GoSleep("3s", func() { m.Cmd("", PROXY_CONF, kit.Select(ice.Info.NodeName, m.Option("nodename"))) })
				} else if runtime.GOOS == cli.WINDOWS {
					m.Cmd(cli.SYSTEM, cli.ECHO, "-ne", kit.Format("\033]0;%s %s serve start %s\007",
						path.Base(kit.Path("")), strings.TrimPrefix(kit.Path(os.Args[0]), kit.Path("")+nfs.PS), kit.JoinCmdArgs(arg...)))
				}
				_serve_start(m)
			}},
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(ice.DEMO) == ice.TRUE, func() { m.Cmd(CHAT_HEADER, ice.DEMO) })
				kit.If(os.Getenv(cli.TERM), func() { m.Go(func() { ssh.PrintQRCode(m, tcp.PublishLocalhost(m, _serve_address(m))) }) })
				m.Cmd(SPIDE, mdb.CREATE, HostPort(m, tcp.LOCALHOST, m.Option(tcp.PORT)), ice.OPS, ice.SRC_MAIN_ICO, nfs.REPOS, "")
				m.Cmds(SPIDE).Table(func(value ice.Maps) {
					kit.If(value[CLIENT_NAME] != ice.OPS && value[TOKEN] != "", func() {
						m.Cmd(SPACE, tcp.DIAL, ice.DEV, value[CLIENT_NAME], TOKEN, value[TOKEN], mdb.TYPE, SERVER)
					})
				})
				Count(m, m.ActionKey(), m.Option(tcp.PORT))
				if cb, ok := m.Optionv(SERVE_START).(func()); ok {
					cb()
				}
				ice.Info.Important = ice.HasVar()
			}},
			PROXY_CONF: {Name: "proxyConf name* port host path", Hand: func(m *ice.Message, arg ...string) {
				if dir := m.OptionDefault(nfs.PATH, PROXY_PATH, tcp.HOST, "127.0.0.1", tcp.PORT, tcp.PORT_9020); true || nfs.Exists(m, dir) {
					for _, p := range []string{"server.conf", "location.conf", "upstream.conf"} {
						m.Cmd(nfs.SAVE, kit.Format("%s/conf/portal/%s/%s", dir, m.Option(mdb.NAME), p), m.Template(p)+lex.NL)
					}
					m.Cmd(cli.SYSTEM, "sudo", kit.Path("usr/local/daemon/10000/sbin/nginx"), "-p", kit.Path("usr/local/daemon/10000/"), "-s", "reload")
				}
			}},
		}, gdb.EventsAction(SERVE_START), mdb.HashAction(
			mdb.SHORT, tcp.PORT, mdb.FIELD, "time,status,port,host,proto"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Action().StatusTimeCount(kit.Dict(ice.MAIN, mdb.Config(m, ice.MAIN)))
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

func ServeCmdAction() ice.Actions {
	return ice.MergeActions(ice.Actions{
		nfs.PS: {Hand: func(m *ice.Message, arg ...string) { RenderCmd(m, "", arg) }},
	})
}
func IsLocalHost(m *ice.Message) bool {
	return (m.R == nil || m.R.Header.Get(html.XForwardedFor) == "") && tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP))
}
func ParseURL(m *ice.Message, p string) []string {
	if u, e := url.Parse(p); e == nil {
		arg := strings.Split(strings.TrimPrefix(u.Path, nfs.PS), nfs.PS)
		for i := 0; i < len(arg); i += 2 {
			switch arg[i] {
			case "s":
				m.Option(ice.POD, kit.Select("", arg, i+1))
			case "c":
				m.Option(ice.CMD, kit.Select("", arg, i+1))
			}
		}
		kit.For(u.Query(), func(k string, v []string) { m.Optionv(k, v) })
		return kit.Split(u.Fragment, ":")
	}
	return []string{}
}
func ParseUA(m *ice.Message) (res []string) {
	res = append(res, aaa.USERROLE, m.Option(ice.MSG_USERROLE))
	res = append(res, aaa.USERNAME, m.Option(ice.MSG_USERNAME))
	res = append(res, aaa.USERNICK, m.Option(ice.MSG_USERNICK))
	res = append(res, aaa.AVATAR, m.Option(ice.MSG_AVATAR))
	res = append(res, cli.DAEMON, m.Option(ice.MSG_DAEMON))
	for _, p := range html.AgentList {
		if strings.Contains(m.Option(ice.MSG_USERUA), p) {
			res = append(res, mdb.ICONS, kit.Select(agentIcons[p], m.Option(mdb.ICONS)), AGENT, p)
			break
		}
	}
	for _, p := range html.SystemList {
		if strings.Contains(m.Option(ice.MSG_USERUA), p) {
			res = append(res, cli.SYSTEM, p)
			break
		}
	}
	return append(res, aaa.IP, m.Option(ice.MSG_USERIP), aaa.UA, m.Option(ice.MSG_USERUA))
}
func ProxyDomain(m *ice.Message, name string) (domain string) {
	p := path.Join(PROXY_PATH, "conf/portal", name, "server.conf")
	if !nfs.Exists(m, p) {
		domain := UserWeb(m).Hostname()
		if domain == "localhost" {
			return ""
		}
		if p = path.Join(PROXY_PATH, "conf/server", kit.Keys(name, kit.Slice(kit.Split(UserWeb(m).Hostname(), "."), -2))) + ".conf"; !nfs.Exists(m, p) {
			return ""
		}
	}
	m.Cmd(nfs.CAT, p, func(ls []string) { kit.If(ls[0] == "server_name", func() { domain = ls[1] }) })
	return kit.Select("", "https://"+domain, domain != "")
}
func Script(m *ice.Message, str string, arg ...ice.Any) string {
	return ice.Render(m, ice.RENDER_SCRIPT, kit.Format(str, arg...))
}
func ChatCmdPath(m *ice.Message, arg ...string) string {
	return m.MergePodCmd("", kit.Select(m.ShortKey(), path.Join(arg...)))
}
func RequireFile(m *ice.Message, file string) string {
	if strings.HasPrefix(file, nfs.PS) || strings.HasPrefix(file, ice.HTTP) || strings.Contains(file, "://") {
		return file
	} else if file != "" {
		return nfs.P + file
	}
	return ""
}
