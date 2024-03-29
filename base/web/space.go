package web

import (
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

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
	"shylinux.com/x/icebergs/misc/websocket"
	kit "shylinux.com/x/toolkits"
)

func _space_qrcode(m *ice.Message, dev string) {
	ssh.PrintQRCode(m, m.Cmdv(SPACE, dev, cli.PWD, mdb.LINK))
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	u := kit.ParseURL(kit.MergeURL2(strings.Replace(m.Cmdv(SPIDE, dev, CLIENT_ORIGIN), HTTP, "ws", 1), PP(SPACE), mdb.TYPE, ice.Info.NodeType, mdb.NAME, name,
		nfs.MODULE, ice.Info.Make.Module, nfs.VERSION, ice.Info.Make.Versions(), arg))
	args := kit.SimpleKV("type,name,host,port", u.Scheme, dev, u.Hostname(), kit.Select(kit.Select("443", "80", u.Scheme == "ws"), u.Port()))
	gdb.Go(m, func() {
		once := sync.Once{}
		redial := kit.Dict(mdb.Configv(m, REDIAL))
		a, b, c := kit.Int(redial["a"]), kit.Int(redial["b"]), kit.Int(redial["c"])
		for i := 1; i < c; i++ {
			next := time.Duration(rand.Intn(a*(i+1))+b*i) * time.Millisecond
			m.Cmd(tcp.CLIENT, tcp.DIAL, args, func(c net.Conn) {
				if c, e := websocket.NewClient(c, u); !m.Warn(e, tcp.DIAL, dev, SPACE, u.String()) {
					defer mdb.HashCreateDeferRemove(m, kit.SimpleKV("", MASTER, dev, u.Host), kit.Dict(mdb.TARGET, c))()
					kit.If(ice.Info.Colors, func() { once.Do(func() { m.Go(func() { _space_qrcode(m, dev) }) }) })
					_space_handle(m.Spawn(), true, dev, c)
					i = 0
				}
			}).Cost(mdb.COUNT, i, mdb.NEXT, next, tcp.DIAL, dev, LINK, u.String()).Sleep(next)
		}
	}, kit.Join(kit.Simple(SPACE, name), lex.SP))
}
func _space_fork(m *ice.Message) {
	addr := kit.Select(m.R.RemoteAddr, m.R.Header.Get(ice.MSG_USERADDR))
	name := kit.ReplaceAll(kit.Select(addr, m.Option(mdb.NAME)), "[", "_", "]", "_", nfs.DF, "_", nfs.PT, "_")
	text := kit.Select(addr, m.Option(mdb.TEXT))
	if kit.IsIn(m.Option(mdb.TYPE), CHROME) && m.Option(mdb.NAME) != CHROME || !(ice.Info.Localhost && tcp.IsLocalHost(m, m.R.RemoteAddr) ||
		m.Option(TOKEN) != "" && m.Cmdv(TOKEN, m.Option(TOKEN), mdb.TIME) > m.Time()) {
		name, text = kit.Hashs(name), kit.Select(addr, m.Option(mdb.NAME), m.Option(mdb.TEXT))
	}
	args := kit.Simple(mdb.TYPE, kit.Select(WORKER, m.Option(mdb.TYPE)), mdb.NAME, name, mdb.TEXT, text, m.OptionSimple(cli.DAEMON, ice.MSG_USERUA), m.OptionSimple(nfs.MODULE, nfs.VERSION))
	if c, e := websocket.Upgrade(m.W, m.R); !m.Warn(e) {
		gdb.Go(m, func() {
			defer mdb.HashCreateDeferRemove(m, args, kit.Dict(mdb.TARGET, c))()
			switch m.Option(mdb.TYPE) {
			case SERVER:
			case WORKER:
				defer gdb.EventDeferEvent(m, DREAM_OPEN, args)(DREAM_CLOSE, args)
			case CHROME:
				m.Go(func() { m.Cmd(SPACE, name, cli.PWD, name) })
			case LOGIN:
				gdb.Event(m, SPACE_LOGIN, args)
			}
			_space_handle(m, false, name, c)
		}, kit.Join(kit.Simple(SPACE, name), lex.SP))
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
		if safe { // 下行权限
			msg.OptionDefault(ice.MSG_USERROLE, aaa.UserRole(msg, msg.Option(ice.MSG_USERNAME)))
		} else { // 上行权限
			kit.If(msg.Option(ice.MSG_USERROLE), func() { msg.Option(ice.MSG_USERROLE, aaa.VOID) })
		}
		source, target := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name), kit.Simple(msg.Optionv(ice.MSG_TARGET))
		msg.Log(tcp.RECV, "%v->%v %v %v", source, target, msg.Detailv(), msg.FormatsMeta(nil))
		if next := msg.Option(ice.MSG_TARGET); next == "" || len(target) == 0 {
			m.Go(func() { _space_exec(msg, source, target, c) }, strings.Join(kit.Simple(SPACE, name, msg.Detailv()), lex.SP))
		} else {
			m.Warn(!mdb.HashSelectDetail(m, next, func(value ice.Map) {
				switch c := value[mdb.TARGET].(type) {
				case (*websocket.Conn): // 转发报文
					_space_echo(msg, source, target, c)
				case ice.Handler: // 接收响应
					m.Go(func() { c(msg) })
				}
			}), ice.ErrNotFound, next)
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
		func() string { return Domain(m.Cmdv(tcp.HOST, aaa.IP), m.Cmdv(SERVE, tcp.PORT)) })
}
func _space_exec(m *ice.Message, source, target []string, c *websocket.Conn) {
	switch kit.Select(cli.PWD, m.Detailv(), 0) {
	case cli.PWD:
		m.Push(mdb.LINK, m.MergePod(kit.Select("", source, -1)))
	default:
		m.Options("__target", kit.Reverse(kit.Simple(source))).OptionDefault(ice.MSG_COUNT, "0")
		kit.If(aaa.Right(m, m.Detailv()), func() { m.TryCatch(m, true, func(_ *ice.Message) { m = m.Cmd() }) })
	}
	defer m.Cost(kit.Format("%v->%v %v %v", source, target, m.Detailv(), m.FormatSize()))
	_space_echo(m.Set(ice.MSG_OPTS).Options(log.DEBUG, m.Option(log.DEBUG)), []string{}, kit.Reverse(kit.Simple(source)), c)
}
func _space_echo(m *ice.Message, source, target []string, c *websocket.Conn) {
	if m.Options(ice.MSG_SOURCE, source, ice.MSG_TARGET, target[1:]); !m.Warn(c.WriteMessage(1, []byte(m.FormatMeta()))) {
		m.Log(tcp.SEND, "%v->%v %v %v", source, target, m.Detailv(), m.FormatsMeta(nil))
	}
}
func _space_send(m *ice.Message, name string, arg ...string) (h string) {
	wait, done := m.Wait(func(msg *ice.Message, arg ...string) {
		m.Cost(kit.Format("%v->[%v] %v %v", m.Optionv(ice.MSG_SOURCE), name, m.Detailv(), msg.FormatSize())).Copy(msg)
	})
	h = mdb.HashCreate(m.Spawn(), mdb.TYPE, tcp.SEND, mdb.NAME, kit.Keys(name, m.Target().ID()), mdb.TEXT, kit.Join(arg, lex.SP), kit.Dict(mdb.TARGET, done))
	defer mdb.HashRemove(m.Spawn(), mdb.HASH, h)
	if target := kit.Split(name, nfs.PT, nfs.PT); mdb.HashSelectDetail(m, target[0], func(value ice.Map) {
		if c, ok := value[mdb.TARGET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotValid, mdb.TARGET) {
			kit.For([]string{ice.MSG_USERROLE}, func(k string) { m.Optionv(k, m.Optionv(k)) })
			kit.For(m.Optionv(ice.MSG_OPTS), func(k string) { m.Optionv(k, m.Optionv(k)) })
			_space_echo(m.Set(ice.MSG_DETAIL, arg...), []string{h}, target, c)
		}
	}) {
		wait()
	} else {
		m.Warn(kit.IndexOf([]string{ice.OPS, ice.DEV}, target[0]) == -1, ice.ErrNotFound, SPACE, name)
	}
	return
}

const (
	CHROME = "chrome"
	MASTER = "master"
	SERVER = "server"
	WORKER = "worker"

	REDIAL = "redial"
)
const (
	SPACE_LOGIN = "space.login"
)
const SPACE = "space"

func init() {
	Index.MergeCommands(ice.Commands{
		SPACE: {Name: "space name cmds auto", Help: "空间站", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, SPACE, ice.MAIN) }},
			ice.MAIN: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(mdb.Config(m, ice.MAIN), func(cmd string) { RenderPodCmd(m, "", cmd) }, func() {
					m.OptionDefault(nfs.VERSION, RenderVersion(m))
					m.RenderResult(nfs.Template(m, "main.html"))
				})
				m.Optionv(ice.MSG_ARGS, kit.Simple(m.Optionv(ice.MSG_ARGS)))
			}},
			tcp.DIAL: {Name: "dial dev=ops name", Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(ice.DEV), HTTP) {
					m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple(ice.DEV))
					m.Option(ice.DEV, ice.DEV)
				}
				_space_dial(m, m.Option(ice.DEV), kit.Select(ice.Info.NodeName, m.Option(mdb.NAME)), arg...)
			}},
			cli.START: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("", tcp.DIAL, arg) }},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				defer mdb.HashModifyDeferRemove(m, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)()
				m.Cmd("", m.Option(mdb.NAME), ice.EXIT)
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					m.Cmds("", func(value ice.Maps) {
						switch value[mdb.TYPE] {
						case MASTER:
							m.PushSearch(mdb.TEXT, m.Cmdv(SPIDE, value[mdb.NAME], CLIENT_ORIGIN), value)
						case SERVER:
							m.PushSearch(mdb.TEXT, m.MergePod(value[mdb.NAME]), value)
						}
					})
				}
			}},
			DOMAIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_space_domain(m)) }},
			LOGIN: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.MSG_USERUA, m.Cmdv("", kit.Select(m.Option(mdb.NAME), arg, 0), ice.MSG_USERUA))
				m.Cmd("", kit.Select(m.Option(mdb.NAME), arg, 0), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case MASTER:
					ctx.ProcessOpen(m, m.Cmdv(SPIDE, m.Option(mdb.NAME), CLIENT_ORIGIN))
				default:
					ctx.ProcessOpen(m, m.MergePod(m.Option(mdb.NAME), arg))
				}
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) { _space_fork(m) }},
		}, mdb.HashAction(mdb.LIMIT, 1000, mdb.LEAST, 1000, mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text,module,version", ctx.ACTION, OPEN, REDIAL, kit.Dict("a", 3000, "b", 1000, "c", 1000)), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 {
				defer m.StatusTimeCount()
				m.Option(ice.MSG_USERWEB, tcp.PublishLocalhost(m, m.Option(ice.MSG_USERWEB)))
				mdb.HashSelect(m.Spawn(), arg...).Sort("").Table(func(index int, value ice.Maps, field []string) {
					if kit.IsIn(value[mdb.TYPE], CHROME, "send") {
						// return
					}
					if m.Push("", value, kit.Split(mdb.Config(m, mdb.FIELD))); len(arg) > 0 && arg[0] != "" {
						m.Push(mdb.STATUS, value[mdb.STATUS])
					}
					if kit.IsIn(value[mdb.TYPE], SERVER, WORKER) {
						m.Push(mdb.LINK, m.MergePod(value[mdb.NAME]))
					} else if value[mdb.TYPE] == CHROME && value[mdb.NAME] != "chrome" {
						m.Push(mdb.LINK, MergeURL2(m, value[mdb.TEXT]))
					} else {
						m.Push(mdb.LINK, "")
					}
					m.PushButton(kit.Select(OPEN, LOGIN, value[mdb.TYPE] == LOGIN), mdb.REMOVE)
				})
				kit.If(!m.IsCliUA(), func() { m.Cmdy("web.code.publish", "contexts", ice.APP) })
				kit.If(len(arg) == 1, func() { m.EchoIFrame(m.MergePod(arg[0])) })
			} else {
				_space_send(m, arg[0], kit.Simple(kit.Split(arg[1]), arg[2:])...)
			}
		}},
	})
}
func Space(m *ice.Message, arg ice.Any) []string {
	if arg == nil || arg == "" {
		return nil
	}
	return []string{SPACE, kit.Format(arg)}
}
