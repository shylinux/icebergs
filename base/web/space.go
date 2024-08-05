package web

import (
	"io"
	"math/rand"
	"net"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

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
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/misc/websocket"
	kit "shylinux.com/x/toolkits"
)

func _space_qrcode(m *ice.Message, dev string) {
	ssh.PrintQRCode(m, m.Cmdv(SPACE, dev, cli.PWD, mdb.LINK))
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	msg := m.Cmd(SPIDE, dev)
	origin := msg.Append(CLIENT_ORIGIN)
	u := kit.ParseURL(kit.MergeURL2(strings.Replace(origin, HTTP, "ws", 1), PP(SPACE), mdb.TYPE, ice.Info.NodeType, mdb.NAME, name, TOKEN, msg.Append(TOKEN), mdb.ICONS, ice.Info.NodeIcon,
		mdb.TIME, ice.Info.Make.Time, nfs.MODULE, ice.Info.Make.Module, nfs.VERSION, ice.Info.Make.Versions(), cli.GOOS, runtime.GOOS, cli.GOARCH, runtime.GOARCH, arg))
	args := kit.SimpleKV("type,name,host,port", u.Scheme, dev, u.Hostname(), kit.Select(kit.Select(tcp.PORT_443, tcp.PORT_80, u.Scheme == "ws"), u.Port()))
	gdb.Go(m, func() {
		once := sync.Once{}
		redial := kit.Dict(mdb.Configv(m, REDIAL))
		a, b, _c := kit.Int(redial["a"]), kit.Int(redial["b"]), kit.Int(redial["c"])
		for i := 1; i < _c; i++ {
			next := time.Duration(rand.Intn(a*i*i)+b*(i+1)) * time.Millisecond
			m.Cmd(tcp.CLIENT, tcp.DIAL, args, func(c net.Conn) {
				if c, e := websocket.NewClient(c, u); !m.WarnNotValid(e, tcp.DIAL, dev, SPACE, u.String()) {
					mdb.HashCreate(m, kit.SimpleKV("", ORIGIN, dev, origin), kit.Dict(mdb.TARGET, c))
					defer mdb.HashRemove(m, mdb.HASH, dev)
					kit.If(ice.Info.Colors, func() { once.Do(func() { m.Go(func() { _space_qrcode(m, dev) }) }) })
					_space_handle(m.Spawn(), true, dev, c)
					i = 0
				}
			}).Cost(mdb.COUNT, i, mdb.NEXT, next, tcp.DIAL, dev, LINK, u.String()).Sleep(next)
			if mdb.HashSelect(m.Spawn(), name).Append(mdb.STATUS) == cli.STOP {
				break
			}
		}
	}, kit.JoinWord(SPACE, dev))
}
func _space_fork(m *ice.Message) {
	addr := kit.Select(m.R.RemoteAddr, m.R.Header.Get(ice.MSG_USERADDR))
	text := strings.ReplaceAll(kit.Select(addr, m.Option(mdb.TEXT)), "%2F", nfs.PS)
	name := SpaceName(kit.Select(addr, m.Option(mdb.NAME)))
	if m.OptionDefault(mdb.TYPE, SERVER) == WORKER && (!nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, name)) || !tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP))) {
		m.Option(mdb.TYPE, SERVER)
	}
	if kit.IsIn(m.Option(mdb.TYPE), PORTAL, aaa.LOGIN) && len(name) == 32 && kit.IsIn(mdb.HashSelects(m.Spawn(), name).Append(aaa.IP), "", m.Option(ice.MSG_USERIP)) {

	} else if kit.IsIn(m.Option(mdb.TYPE), SERVER) && IsLocalHost(m) {

	} else if kit.IsIn(m.Option(mdb.TYPE), WORKER) && IsLocalHost(m) {
		text = nfs.USR_LOCAL_WORK + name
	} else {
		name, text = kit.Hashs(name), kit.Select(addr, m.Option(mdb.TEXT))
	}
	safe := false
	if m.Option(ice.MSG_USERNAME, ""); kit.IsIn(m.Option(mdb.TYPE), WORKER, PORTAL) {
		if tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
			aaa.SessAuth(m, kit.Dict(m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME, ice.Info.Username)).AppendSimple()))
		}
	} else if m.Option(TOKEN) != "" {
		if msg := m.Cmd(TOKEN, m.Option(TOKEN)); msg.Append(mdb.TIME) > m.Time() && kit.IsIn(msg.Append(mdb.TYPE), SERVER, SPIDE) {
			aaa.SessAuth(m, kit.Dict(m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME, msg.Append(mdb.NAME))).AppendSimple()))
			name = SpaceName(kit.Select(name, msg.Append(mdb.TEXT)))
			kit.If(ProxyDomain(m, name), func(p string) { text = p })
			safe = aaa.IsTechOrRoot(m)
		}
	}
	args := kit.Simple(mdb.TYPE, m.Option(mdb.TYPE), mdb.NAME, name, mdb.TEXT, text, m.OptionSimple(mdb.ICONS, mdb.TIME, nfs.MODULE, nfs.VERSION, cli.DAEMON))
	args = append(args, aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
	args = append(args, cli.SYSTEM, m.Option(cli.GOOS))
	args = append(args, ParseUA(m)...)
	if c, e := websocket.Upgrade(m.W, m.R); !m.WarnNotValid(e) {
		gdb.Go(m, func() {
			defer mdb.HashCreateDeferRemove(m, args, kit.Dict(mdb.TARGET, c))()
			switch m.Option(mdb.TYPE) {
			case LOGIN:
				if m.Option(ice.MSG_SESSID) != "" && m.Option(ice.MSG_USERNAME) != "" {
					m.Cmd(SPACE, name, ice.MSG_SESSID, m.Option(ice.MSG_SESSID))
				}
				gdb.Event(m, SPACE_LOGIN, args)
				defer gdb.Event(m, SPACE_LOGIN_CLOSE, args)
			case PORTAL:
				gdb.EventDeferEvent(m, PORTAL_OPEN, args)
				m.Go(func() { m.Cmd(SPACE, name, cli.PWD, name) })
			case WORKER:
				defer gdb.EventDeferEvent(m, DREAM_OPEN, args)(DREAM_CLOSE, args)
				safe = true
				m.Go(func() {
					SpacePwd(m, name, kit.Path(""))
					SpaceEvent(m, OPS_SERVER_OPEN, name, args...)
				})
			case SERVER:
				defer gdb.EventDeferEvent(m, SPACE_OPEN, args)(SPACE_CLOSE, args)
				m.Go(func() {
					SpacePwd(m, name, "")
					SpaceEvent(m.Spawn(ice.MSG_USERROLE, aaa.TECH), OPS_SERVER_OPEN, name, args...)
				})
			}
			_space_handle(m, safe, name, c)
		}, kit.JoinWord(SPACE, name))
	}
}
func _space_handle(m *ice.Message, safe bool, name string, c *websocket.Conn) {
	defer m.Cost(SPACE, name)
	m.Options(ice.MSG_USERROLE, "", mdb.TYPE, "", mdb.NAME, "", cli.DAEMON, "")
	for {
		_, b, e := c.ReadMessage()
		if e != nil {
			break
		}
		msg := m.Spawn(b)
		if safe && msg.Option(ice.MSG_UNSAFE) != ice.TRUE { // 下行权限
			if !aaa.IsTechOrRoot(msg) && msg.Option(ice.MSG_HANDLE) != ice.TRUE {
				msg.Option(ice.MSG_USERROLE, kit.Select(msg.Option(ice.MSG_USERROLE), aaa.UserRole(msg, msg.Option(ice.MSG_USERNAME))))
			}
			// kit.If(kit.IsIn(msg.Option(ice.MSG_USERROLE), "", aaa.VOID), func() { msg.Option(ice.MSG_USERROLE, aaa.UserRole(msg, msg.Option(ice.MSG_USERNAME))) })
		} else { // 上行权限
			msg.Option(ice.MSG_UNSAFE, ice.TRUE)
			kit.If(msg.Option(ice.MSG_USERROLE), func() { msg.Option(ice.MSG_USERROLE, aaa.VOID) })
		}
		source, target := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name), kit.Simple(msg.Optionv(ice.MSG_TARGET))
		msg.Log(kit.Select(tcp.RECV, tcp.ECHO, msg.Option(ice.MSG_HANDLE) == ice.TRUE), "%v->%v %v %v", source, target, msg.Detailv(), msg.FormatMeta())
		if next := msg.Option(ice.MSG_TARGET); next == "" || len(target) == 0 {
			msg.Go(func() {
				if k := kit.Keys(msg.Option(ice.MSG_USERPOD), "_token"); msg.Option(k) != "" {
					aaa.SessCheck(msg, msg.Option(k))
				}
				msg.Option(ice.MSG_OPTS, kit.Simple(msg.Optionv(ice.MSG_OPTION), func(k string) bool { return !strings.HasPrefix(k, ice.MSG_SESSID) }))
				_space_exec(msg, name, source, target, c)
			}, strings.Join(kit.Simple(SPACE, name, msg.Detailv()), lex.SP))
		} else {
			for i := 0; i < 5; i++ {
				if !m.WarnNotFoundSpace(!mdb.HashSelectDetail(m, next, func(value ice.Map) {
					switch c := value[mdb.TARGET].(type) {
					case (*websocket.Conn): // 转发报文
						kit.If(value[mdb.TYPE] == ORIGIN && msg.Option(ice.MSG_HANDLE) == ice.FALSE, func() {
							msg.Optionv(ice.MSG_USERWEB, kit.Simple(value[mdb.TEXT], msg.Optionv(ice.MSG_USERWEB)))
							msg.Optionv(ice.MSG_USERPOD, kit.Simple(kit.Keys(target[1:]), msg.Optionv(ice.MSG_USERPOD)))
						})
						_space_echo(msg, source, target, c)
					case ice.Handler: // 接收响应
						msg.Go(func() { c(msg) })
					}
				}), SPACE, next) {
					break
				}
				m.Info("what %v", msg.FormatStack(1, 100))
				m.Info("what %v", msg.FormatChain())
				if kit.HasPrefixList(msg.Detailv(), "toast") {
					break
				}
				if msg.Option("space.noecho") == "true" {
					break
				}
				m.Sleep3s()
			}
		}
	}
}
func _space_domain(m *ice.Message) (link string) {
	return kit.GetValid(
		func() string { return ice.Info.Domain },
		func() string {
			if dev := kit.Select(ice.DEV, ice.OPS, ice.Info.NodeType == WORKER); mdb.HashSelectDetail(m, dev, nil) {
				m.Options(ice.MSG_OPTION, ice.MSG_USERNAME, ice.MSG_OPTS, ice.MSG_USERNAME)
				return m.Cmdv(SPACE, dev, cli.PWD, mdb.LINK)
			}
			return ""
		},
		func() string { return tcp.PublishLocalhost(m, m.Option(ice.MSG_USERWEB)) },
		func() string { return HostPort(m, m.Cmdv(tcp.HOST, aaa.IP), m.Cmdv(SERVE, tcp.PORT)) },
	)
}
func _space_exec(m *ice.Message, name string, source, target []string, c *websocket.Conn) {
	m.Option(ice.MSG_HANDLE, ice.TRUE)
	switch kit.Select("", m.Detailv(), 0) {
	case "":
		m.Warn(true, ice.ErrNotValid)
		return
	case cli.PWD:
		m.Push(mdb.LINK, m.MergePod(kit.Select("", source, -1)))
		if m.Option(cli.SYSTEM) == "" {
			m.Optionv(ice.MSG_OPTION, []string{})
			break
		}
		args := m.OptionSimple(mdb.ICONS, mdb.TIME, nfs.MODULE, nfs.VERSION, AGENT, cli.SYSTEM)
		kit.If(name == ice.OPS, func() { args = append(args, m.OptionSimple(mdb.TEXT)...) })
		mdb.HashModify(m, mdb.HASH, name, ParseUA(m), args)
		SpaceEvent(m, OPS_ORIGIN_OPEN, name, kit.Simple(mdb.NAME, name, args)...)
	default:
		if m.IsErr() {
			return
		}
		m.Options(ice.MSG_ARGS, "", ice.MSG_COUNT, "0")
		kit.If(m.Option(ice.MSG_DAEMON), func(p string) {
			m.Option(ice.MSG_DAEMON0, m.Option(ice.MSG_DAEMON))
			m.Option(ice.MSG_DAEMON, kit.Keys(kit.Slice(kit.Reverse(kit.Simple(source)), 0, -1), p))
		})
		m.Option(ice.FROM_SPACE, kit.Keys(kit.Reverse(kit.Simple(source[1:]))))
		kit.If(aaa.Right(m, m.Detailv()), func() { m.TryCatch(true, func(_ *ice.Message) { m = m.Cmd() }) })
		kit.If(m.Optionv(ice.MSG_ARGS) != nil, func() { m.Options(ice.MSG_ARGS, kit.Simple(m.Optionv(ice.MSG_ARGS))) })
	}
	defer m.Cost(kit.Format("%v->%v %v %v", source, target, m.Detailv(), m.FormatSize()))
	if m.Option(ice.SPACE_NOECHO) == ice.TRUE {
		return
	}
	m.Options(ice.MSG_USERWEB, m.Optionv(ice.MSG_USERWEB), ice.MSG_USERPOD, m.Optionv(ice.MSG_USERPOD))
	_space_echo(m.Set(ice.MSG_OPTS).Options(m.OptionSimple(ice.MSG_HANDLE, ice.LOG_DEBUG, ice.LOG_DISABLE, ice.LOG_TRACEID)), []string{}, kit.Reverse(kit.Simple(source)), c)
}
func _space_echo(m *ice.Message, source, target []string, c *websocket.Conn) {
	defer func() { m.WarnNotValid(recover()) }()
	if m.Options(ice.MSG_SOURCE, source, ice.MSG_TARGET, target[1:]); !m.WarnNotValid(c.WriteMessage(1, []byte(m.FormatMeta()))) {
		if source != nil {
			m.Log(kit.Select(tcp.SEND, tcp.DONE, m.Option(ice.MSG_HANDLE) == ice.TRUE), "%v->%v %v %v", source, target, kit.ReplaceAll(kit.Format("%v", m.Detailv()), "\r\n", "\\r\\n", "\t", "\\t", "\n", "\\n"), m.FormatMeta())
		}
	}
}
func _space_send(m *ice.Message, name string, arg ...string) (h string) {
	withecho := m.Option(ice.SPACE_NOECHO) != ice.TRUE
	kit.If(len(arg) > 0 && arg[0] == TOAST, func() { withecho = false; m.Option(ice.MSG_DEBUG, ice.FALSE) })
	wait, done := m.Wait(kit.Select("", m.OptionDefault(ice.SPACE_TIMEOUT, "180s"), withecho), func(msg *ice.Message, arg ...string) {
		m.Cost(kit.Format("%v->[%v] %v %v", m.Optionv(ice.MSG_SOURCE), name, m.Detailv(), msg.FormatSize())).Copy(msg)
	})
	if withecho {
		h = mdb.HashCreate(m.SpawnSilent(), mdb.TYPE, tcp.SEND, mdb.NAME, kit.Keys(name, m.Target().ID()), mdb.TEXT, kit.Join(arg, lex.SP), kit.Dict(mdb.TARGET, done))
		defer mdb.HashRemove(m.SpawnSilent(), mdb.HASH, h)
	}
	if target := kit.Split(name, nfs.PT, nfs.PT); !mdb.HashSelectDetail(m, target[0], func(value ice.Map) {
		if c, ok := value[mdb.TARGET].(*websocket.Conn); !m.WarnNotValid(!ok, mdb.TARGET) {
			kit.If(kit.Format(value[mdb.TYPE]) == ORIGIN && target[0] != ice.OPS, func() {
				m.Optionv(ice.MSG_USERWEB, kit.Simple(value[mdb.TEXT], m.Optionv(ice.MSG_USERWEB)))
				m.Optionv(ice.MSG_USERPOD, kit.Simple(kit.Keys(target[1:]), m.Optionv(ice.MSG_USERPOD)))
				m.Options(ice.MSG_USERHOST, "", ice.MSG_USERWEB0, m.Option(ice.MSG_USERWEB), ice.MSG_USERPOD0, name)
			})
			m.Option(ice.MSG_HANDLE, ice.FALSE)
			kit.For([]string{ice.MSG_USERROLE, ice.LOG_TRACEID, ice.SPACE_NOECHO}, func(k string) { m.Optionv(k, m.Optionv(k)) })
			kit.For(kit.Filters(kit.Simple(m.Optionv(ice.MSG_OPTS)), "task.id", "work.id"), func(k string) { m.Optionv(k, m.Optionv(k)) })
			if withecho {
				_space_echo(m.Set(ice.MSG_DETAIL, arg...), []string{h}, target, c)
			} else {
				_space_echo(m.Set(ice.MSG_DETAIL, arg...), nil, target, c)
			}
		}
	}) {
		if name == ice.OPS && ice.Info.NodeType == SERVER {
			m.Cmdy(arg)
			return
		}
		kit.If(m.IsDebug(), func() {
			m.WarnNotFoundSpace(kit.IndexOf([]string{ice.OPS, ice.DEV}, target[0]) == -1, SPACE, name)
		})
	} else if withecho {
		m.Warn(!wait(), kit.Format("space %v %v time out", name, arg))
	}
	return
}

const (
	WEIXIN = "weixin"
	PORTAL = "portal"
	WORKER = "worker"
	SERVER = "server"
	MYSELF = "myself"
	ORIGIN = "origin"

	REDIAL = "redial"
	AGENT  = "agent"
)
const (
	OPS_ORIGIN_OPEN = "ops.origin.open"
	OPS_SERVER_OPEN = "ops.server.open"
	OPS_DREAM_SPAWN = "ops.dream.spawn"

	SPACE_LOGIN       = "space.login"
	SPACE_LOGIN_CLOSE = "space.login.close"
	SPACE_GRANT       = "space.grant"
	SPACE_OPEN        = "space.open"
	SPACE_CLOSE       = "space.close"
	PORTAL_OPEN       = "portal.open"
	PORTAL_CLOSE      = "portal.close"
)
const SPACE = "space"

func init() {
	Index.MergeCommands(ice.Commands{
		"p": {Help: "资源", Actions: ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			if kit.IsIn(arg[0], ice.SRC, ice.USR) {
				ShareLocalFile(m, arg...)
			} else {
				m.Cmdy(PP(ice.REQUIRE), arg)
			}
		}},
		"m": {Help: "模块", Actions: ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			p := path.Join(nfs.USR_MODULES, path.Join(arg...))
			kit.If(!nfs.Exists(m, p), func() {
				if kit.IsIn(m.Option(ice.MSG_USERROLE), aaa.TECH, aaa.ROOT) {
					kit.If(!nfs.Exists(m, nfs.USR_PACKAGE), func() {
						m.Cmd(nfs.SAVE, nfs.USR_PACKAGE, kit.Formats(kit.Dict(mdb.NAME, "usr", nfs.VERSION, "0.0.1")))
					})
					m.Cmd(cli.SYSTEM, "npm", "install", arg[0], kit.Dict(cli.CMD_DIR, ice.USR))
				}
			})
			m.RenderDownload(p)
		}},
		"c": {Help: "命令", Actions: ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) { m.Cmdy(CHAT_CMD, arg) }},
		"s": {Help: "空间", Actions: ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) { m.Cmdy(CHAT_POD, arg) }},
		SPACE: {Name: "space name cmds auto", Help: "空间站", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				cli.NodeInfo(m, ice.Info.Pathname, WORKER)
				aaa.White(m, SPACE, ice.MAIN)
				if kit.IsIn(ice.Info.NodeIcon, "src/main.ico", "") {
					nfs.Exists(m, "src/main.ico", func(p string) { ice.Info.NodeIcon = p })
					nfs.Exists(m, "src/main.png", func(p string) { ice.Info.NodeIcon = p })
					nfs.Exists(m, "src/main.jpg", func(p string) { ice.Info.NodeIcon = p })
				}
			}},
			mdb.ICONS: {Hand: func(m *ice.Message, arg ...string) {
				cli.NodeInfo(m, ice.Info.Pathname, WORKER, arg[0])
				m.Cmd(SERVE, m.ActionKey(), arg)
			}},
			ice.MAIN: {Name: "main index", Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					ice.Info.NodeMain = m.Option(ctx.INDEX)
					m.Cmd(SERVE, m.ActionKey(), arg)
					return
				}
				m.Options(mdb.ICONS, "")
				kit.If(ice.Info.NodeMain, func(cmd string) { RenderPodCmd(m, "", cmd) }, func() { RenderMain(m) })
			}},
			ice.INFO: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				m.Push(mdb.TIME, ice.Info.Make.Time)
				m.Push(mdb.NAME, ice.Info.NodeName)
				m.Push(mdb.ICONS, ice.Info.NodeIcon)
				m.Push(nfs.MODULE, ice.Info.Make.Module)
				m.Push(nfs.VERSION, ice.Info.Make.Versions())
				m.Push(nfs.PATHNAME, ice.Info.Pathname)
				m.Push(tcp.HOSTPORT, HostPort(m, m.Cmd(tcp.HOST).Append(aaa.IP), m.Cmd(SERVER).Append(tcp.PORT)))
				m.Push(ORIGIN, m.Option(ice.MSG_USERHOST))
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					m.Cmds("", func(value ice.Maps) {
						switch value[mdb.TYPE] {
						case SERVER:
							m.PushSearch(mdb.TEXT, m.MergePod(value[mdb.NAME]), value)
						case ORIGIN:
							m.PushSearch(mdb.TEXT, m.Cmdv(SPIDE, value[mdb.NAME], CLIENT_ORIGIN), value)
						}
					})
				}
			}},
			cli.START: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("", tcp.DIAL, arg) }},
			tcp.DIAL: {Name: "dial dev=ops name", Hand: func(m *ice.Message, arg ...string) {
				_space_dial(m, m.Option(ice.DEV), kit.Select(ice.Info.NodeName, m.Option(mdb.NAME)), arg...)
				ice.Info.Important = ice.HasVar()
			}},
			cli.CLOSE: {Hand: func(m *ice.Message, arg ...string) { mdb.HashRemove(m, m.OptionSimple(mdb.NAME)) }},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				defer ToastProcess(m)()
				mdb.HashModify(m, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				msg := mdb.HashSelect(m.Spawn(), m.Option(mdb.NAME))
				if msg.Append(mdb.TYPE) == ORIGIN {
					if target, ok := mdb.HashSelectTarget(m, m.Option(mdb.NAME), nil).(io.Closer); ok {
						target.Close()
					}
				} else {
					m.Cmd("", m.Option(mdb.NAME), ice.EXIT).Sleep3s()
				}
			}},
			DOMAIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_space_domain(m)) }},
			LOGIN: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd("", kit.Select(m.Option(mdb.NAME), arg, 0))
				m.Options(ice.MSG_USERIP, msg.Append(aaa.IP), ice.MSG_USERUA, msg.Append(aaa.UA))
				m.Cmd("", kit.Select(m.Option(mdb.NAME), arg, 0), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
			}},
			SPACE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(ice.FROM_DAEMON), func(p string) { m.Cmd("", p, GRANT, m.Option(mdb.NAME), -1) })
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case ORIGIN:
					ProcessIframe(m, m.Option(mdb.NAME), SpideOrigin(m, m.Option(mdb.NAME)), arg...)
				default:
					ProcessIframe(m, m.Option(mdb.NAME), m.MergePod(m.Option(mdb.NAME)), arg...)
				}
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) { _space_fork(m) }},
		}, gdb.EventsAction(SPACE_LOGIN), mdb.HashAction(mdb.LIMIT, 1000, mdb.LEAST, 500,
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text,icons,module,version,agent,system,ip,usernick,username,userrole",
			ctx.ACTION, OPEN, REDIAL, kit.Dict("a", 1000, "b", 100, "c", 1000),
		), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 {
				if len(arg) == 1 && strings.Contains(arg[0], nfs.PT) {
					ls := kit.Split(arg[0], nfs.PT)
					m.Cmdy(SPACE, ls[0], SPACE, kit.Keys(ls[1:]))
					return
				}
				defer m.StatusTimeCount(kit.Dict(ice.MAIN, mdb.Config(m, ice.MAIN)))
				kit.If(len(arg) > 0 && arg[0] != "", func() { m.OptionFields(ice.MSG_DETAIL) })
				mdb.HashSelect(m.Spawn(), arg...).Table(func(value ice.Maps) {
					if m.Push("", value, kit.Split(mdb.Config(m, mdb.FIELD))); len(arg) > 0 && arg[0] != "" {
						m.Push(mdb.STATUS, value[mdb.STATUS]).Push(aaa.UA, value[aaa.UA])
					}
					if kit.IsIn(value[mdb.TYPE], WEIXIN, PORTAL) && value[mdb.NAME] != html.CHROME {
						m.Push(mdb.LINK, m.MergeLink(value[mdb.TEXT]))
					} else if kit.IsIn(value[mdb.TYPE], WORKER, SERVER) {
						m.Push(mdb.LINK, m.MergePod(value[mdb.NAME]))
					} else if kit.IsIn(value[mdb.TYPE], ORIGIN) {
						m.Push(mdb.LINK, value[mdb.TEXT])
					} else {
						m.Push(mdb.LINK, "")
					}
					m.PushButton(kit.Select(OPEN, LOGIN, value[mdb.TYPE] == LOGIN), mdb.REMOVE)
				})
				m.RewriteAppend(func(value, key string, index int) string {
					if key == mdb.ICONS && !kit.HasPrefix(value, HTTP) {
						if m.Appendv(mdb.TYPE)[index] == ORIGIN {
							if !kit.HasPrefix(value, nfs.PS) {
								value = kit.MergeURL(nfs.P + value)
							}
							if m.Appendv(mdb.NAME)[index] == ice.OPS {
								value = kit.MergeURL2(m.Option(ice.MSG_USERWEB), value)
							} else {
								value = kit.MergeURL2(m.Appendv(mdb.TEXT)[index], value)
							}
						} else {
							if !kit.HasPrefix(value, nfs.PS) {
								value = kit.MergeURL(nfs.P+value, ice.POD, kit.Keys(m.Option(ice.MSG_USERPOD), m.Appendv(mdb.NAME)[index]))
							}
						}
					}
					return value
				})
				m.Sort("", kit.Simple(aaa.LOGIN, WEIXIN, PORTAL, WORKER, SERVER, ORIGIN))
			} else {
				if kit.IsIn(arg[0], "", ice.CONTEXTS) {
					m.Cmdy(arg[1:])
					return
				}
				for i := 0; i < 5; i++ {
					if _space_send(m, arg[0], kit.Simple(kit.Split(arg[1]), arg[2:])...); !m.IsErrNotFoundSpace() {
						break
					} else if i < 4 {
						m.SetAppend().SetResult().Sleep3s()
					}
				}
			}
		}},
	})
	ice.Info.Inputs = append(ice.Info.Inputs, func(m *ice.Message, arg ...string) {
		switch kit.TrimPrefix(arg[0], "extra.") {
		case DREAM:
			m.SetAppend()
			AdminCmd(m, DREAM).Table(func(value ice.Maps) {
				kit.If(kit.IsIn(value[mdb.TYPE], WORKER), func() {
					m.Push(arg[0], value[mdb.NAME])
					m.PushRecord(value, nfs.VERSION, mdb.TIME, nfs.MODULE, mdb.ICONS)
				})
			})
		case SPACE:
			AdminCmd(m, SPACE).Table(func(value ice.Maps) {
				kit.If(kit.IsIn(value[mdb.TYPE], WORKER, SERVER), func() { m.Push(arg[0], value[mdb.NAME]) })
			})
		case SERVER:
			AdminCmd(m, SPACE).Table(func(value ice.Maps) {
				kit.If(kit.IsIn(value[mdb.TYPE], SERVER), func() { m.Push(arg[0], value[mdb.NAME]) })
			})
		case ORIGIN:
			m.SetAppend().Push(arg[0], SpideOrigin(m, ice.DEV))
			m.Copy(m.Cmd(SPIDE, kit.Dict(ice.MSG_FIELDS, CLIENT_ORIGIN)).CutTo(CLIENT_ORIGIN, arg[0]).Sort(arg[0]))
		case mdb.ICONS:
			m.Options(nfs.DIR_REG, kit.ExtReg(nfs.PNG, nfs.JPG, nfs.JPEG), nfs.DIR_DEEP, ice.TRUE)
			m.Cmdy(nfs.DIR, nfs.SRC, nfs.PATH)
			m.Cmdy(nfs.DIR, ice.USR_LOCAL_IMAGE, nfs.PATH)
			m.Cmdy(nfs.DIR, ice.USR_ICONS, nfs.PATH)
			m.CutTo(nfs.PATH, arg[0])
		case ctx.INDEX, ice.CMD:
			m.OptionFields(ctx.INDEX)
			if space := m.Option(SPACE); space != "" {
				m.Options(SPACE, []string{}).Cmdy(SPACE, space, ctx.COMMAND)
			} else {
				m.Cmdy(ctx.COMMAND)
			}
			m.CutTo(ctx.INDEX, arg[0])
		case ctx.ARGS:
			m.OptionDefault(ctx.INDEX, m.Option("extra.index"))
			if space := m.Option(SPACE); space != "" {
				m.Options(SPACE, []string{}).Cmdy(SPACE, space, ctx.COMMAND, mdb.INPUTS, m.Option(ctx.INDEX))
			} else {
				m.Cmdy(ctx.COMMAND, mdb.INPUTS, m.Option(ctx.INDEX))
			}
		case tcp.WIFI:
			m.Cmdy(tcp.WIFI).CutTo(tcp.SSID, arg[0])
		case MESSAGE:
			m.Cmdy(MESSAGE).Cut(mdb.HASH, mdb.ZONE, mdb.ICONS)
		case TARGET:
			m.AdminCmd(MATRIX).Table(func(value ice.Maps) {
				m.Push(arg[0], kit.Keys(kit.Select("", ice.OPS, ice.Info.NodeType == WORKER), value[DOMAIN], value[mdb.NAME]))
				m.Push(mdb.TYPE, value[mdb.TYPE])
			})
			m.Sort("type,target", []string{MYSELF, SERVER, ORIGIN, WORKER}, ice.STR_R)
		}
	})
	ice.Info.AdminCmd = AdminCmd
	ctx.PodCmd = func(m *ice.Message, arg ...ice.Any) bool {
		Upload(m)
		if pod := m.Option(ice.POD); pod != "" {
			if ls := kit.Simple(m.Optionv(ice.MSG_UPLOAD)); len(ls) > 1 {
				m.Cmd(SPACE, pod, SPIDE, ice.DEV, CACHE, SHARE_CACHE+ls[0])
			}
			m.Options(ice.POD, []string{}, ice.MSG_USERPOD, strings.TrimPrefix(pod, "ops.")).Cmdy(append(kit.List(ice.SPACE, pod), arg...)...)
			return true
		}
		return false
	}
}
func Space(m *ice.Message, arg ice.Any) []string {
	if arg == nil || arg == "" {
		return nil
	}
	return []string{SPACE, kit.Format(arg)}
}
func PodCmd(m *ice.Message, key string, arg ...string) bool {
	if pod := m.Option(key); pod != "" {
		m.Options(key, "", ice.MSG_USERPOD, strings.TrimPrefix(pod, "ops.")).Cmdy(SPACE, pod, m.ShortKey(), arg)
		return true
	} else {
		return false
	}
}
func SpaceName(name string) string {
	return kit.ReplaceAll(name, nfs.DF, "_", nfs.PS, "_", nfs.PT, "_", "[", "_", "]", "_")
}
func SpacePwd(m *ice.Message, name, text string) {
	m.Optionv(ice.MSG_OPTS, m.OptionSimple(ice.MSG_USERROLE, ice.MSG_USERNAME))
	m.Cmd(SPACE, name, cli.PWD, name, kit.Dict(ice.SPACE_NOECHO, ice.TRUE,
		mdb.ICONS, ice.Info.NodeIcon, mdb.TEXT, text, AGENT, "Go-http-client", cli.SYSTEM, runtime.GOOS,
		mdb.TIME, ice.Info.Make.Time, nfs.MODULE, ice.Info.Make.Module, nfs.VERSION, ice.Info.Make.Versions(),
	))
}
func SpaceEvent(m *ice.Message, event, skip string, arg ...string) {
	m.Optionv(ice.MSG_OPTS, m.OptionSimple(ice.MSG_USERROLE, ice.MSG_USERNAME))
	m.Cmds(SPACE).Table(func(value ice.Maps) {
		if kit.IsIn(value[mdb.TYPE], WORKER) && value[mdb.NAME] != skip {
			m.Cmd(SPACE, value[mdb.NAME], gdb.EVENT, gdb.HAPPEN, gdb.EVENT, event, arg, kit.Dict(
				ice.MSG_USERROLE, aaa.TECH, ice.SPACE_NOECHO, ice.TRUE,
			))
		}
	})
	m.Cmd(gdb.EVENT, gdb.HAPPEN, gdb.EVENT, event, arg, kit.Dict(ice.MSG_USERROLE, aaa.TECH))
}
