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
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/websocket"
)

func _space_qrcode(m *ice.Message, dev string) {
	ssh.PrintQRCode(m, m.Cmdv(SPACE, dev, cli.PWD, mdb.LINK))
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	uri := kit.ParseURL(kit.MergeURL2(strings.Replace(m.Cmdv(SPIDE, dev, CLIENT_ORIGIN), HTTP, "ws", 1), PP(SPACE), mdb.TYPE, ice.Info.NodeType, mdb.NAME, name, arg))
	args := kit.SimpleKV("type,name,host,port", uri.Scheme, dev, uri.Hostname(), uri.Port())
	m.Go(func() {
		once := sync.Once{}
		redial := kit.Dict(mdb.Configv(m, REDIAL))
		a, b, c := kit.Int(redial["a"]), kit.Int(redial["b"]), kit.Int(redial["c"])
		for i := 1; i < c; i++ {
			next := time.Duration(rand.Intn(a*(i+1))+b*i) * time.Millisecond
			m.Cmd(tcp.CLIENT, tcp.DIAL, args, func(c net.Conn) {
				if c, _, e := websocket.NewClient(c, uri, nil, ice.MOD_BUFS, ice.MOD_BUFS); !m.Warn(e, tcp.DIAL, dev, SPACE, uri.String()) {
					defer mdb.HashCreateDeferRemove(m, kit.SimpleKV("", MASTER, dev, uri.Hostname()), kit.Dict(mdb.TARGET, c))()
					kit.If(ice.Info.Colors, func() { once.Do(func() { m.Go(func() { _space_qrcode(m, dev) }) }) })
					_space_handle(m.Spawn(), true, dev, c)
					i = 0
				}
			}).Cost("order", i, "sleep", next, tcp.DIAL, dev, "uri", uri.String()).Sleep(next)
		}
	}, kit.Join(kit.Simple(SPACE, name), ice.SP))
}
func _space_fork(m *ice.Message) {
	addr := kit.Select(m.R.RemoteAddr, m.R.Header.Get(ice.MSG_USERADDR))
	name := kit.ReplaceAll(kit.Select(addr, m.Option(mdb.NAME)), "[", "_", "]", "_", ice.DF, "_", ice.PT, "_")
	args := kit.Simple(mdb.TYPE, kit.Select(WORKER, m.Option(mdb.TYPE)), mdb.NAME, name, mdb.TEXT, kit.Select(addr, m.Option(mdb.TEXT)), m.OptionSimple(cli.DAEMON, ice.MSG_USERUA))
	if c, e := websocket.Upgrade(m.W, m.R, nil, ice.MOD_BUFS, ice.MOD_BUFS); !m.Warn(e) {
		m.Go(func() {
			defer mdb.HashCreateDeferRemove(m, args, kit.Dict(mdb.TARGET, c))()
			switch m.Option(mdb.TYPE) {
			case WORKER:
				defer gdb.EventDeferEvent(m, DREAM_OPEN, args)(DREAM_CLOSE, args)
			case CHROME:
				m.Go(func() { m.Cmd(SPACE, name, cli.PWD, name) })
			case LOGIN:
				gdb.Event(m, SPACE_LOGIN, args)
			}
			_space_handle(m, false, name, c)
		}, kit.Join(kit.Simple(SPACE, name), ice.SP))
	}
}
func _space_handle(m *ice.Message, safe bool, name string, c *websocket.Conn) {
	defer m.Cost(SPACE, name)
	for {
		_, b, e := c.ReadMessage()
		if e != nil {
			break
		}
		msg := m.Spawn(b)
		source, target := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name), kit.Simple(msg.Optionv(ice.MSG_TARGET))
		msg.Log("recv", "%v->%v %v %v", source, target, msg.Detailv(), msg.FormatsMeta(nil))
		if next := msg.Option(ice.MSG_TARGET); next == "" || len(target) == 0 {
			if safe { // 下行命令
				msg.Option(ice.MSG_USERROLE, aaa.UserRole(msg, msg.Option(ice.MSG_USERNAME)))
			} else { // 上行请求
				msg.Option(ice.MSG_USERROLE, aaa.VOID)
			}
			if msg.Option("_exec") == "go" {
				m.Go(func() { _space_exec(msg, source, target, c) }, strings.Join(kit.Simple(SPACE, name, msg.Detailv()), ice.SP))
			} else {
				_space_exec(msg, source, target, c)
			}
		} else {
			m.Warn(!mdb.HashSelectDetail(m, next, func(value ice.Map) {
				switch c := value[mdb.TARGET].(type) {
				case (*websocket.Conn):
					_space_echo(msg, source, target, c) // 转发报文
				case ice.Handler:
					c(msg) // 接收响应
				}
			}), ice.ErrNotFound, next)
		}
	}
}
func _space_domain(m *ice.Message) (link string) {
	m.Options(ice.MSG_OPTION, ice.MSG_USERNAME, ice.MSG_OPTS, ice.MSG_USERNAME)
	return kit.GetValid(
		func() string { return ice.Info.Domain },
		func() string {
			if dev := kit.Select(ice.DEV, ice.OPS, ice.Info.NodeType == WORKER); mdb.HashSelectDetail(m, dev, nil) {
				return m.Cmdv(SPACE, dev, cli.PWD, mdb.LINK)
			}
			return ""
		},
		func() string { return tcp.PublishLocalhost(m, m.Option(ice.MSG_USERWEB)) },
		func() string { return kit.Format("http://%s:%s", m.Cmdv(tcp.HOST, aaa.IP), m.Cmdv(SERVE, tcp.PORT)) })
}
func _space_exec(m *ice.Message, source, target []string, c *websocket.Conn) {
	switch kit.Select(cli.PWD, m.Detailv(), 0) {
	case cli.PWD:
		m.Push(mdb.LINK, kit.MergePOD(_space_domain(m), kit.Select("", source, -1)))
	default:
		kit.If(aaa.Right(m, m.Detailv()), func() { m = m.Cmd() })
	}
	defer m.Cost(kit.Format("%v->%v %v %v", source, target, m.Detailv(), m.FormatSize()))
	_space_echo(m.Set(ice.MSG_OPTS).Options(log.DEBUG, m.Option(log.DEBUG)), []string{}, kit.Reverse(kit.Simple(source)), c)
}
func _space_echo(m *ice.Message, source, target []string, c *websocket.Conn) {
	if m.Options(ice.MSG_SOURCE, source, ice.MSG_TARGET, target[1:]); !m.Warn(c.WriteMessage(1, []byte(m.FormatMeta()))) {
		m.Log("send", "%v->%v %v %v", source, target, m.Detailv(), m.FormatsMeta(nil))
	}
}
func _space_send(m *ice.Message, name string, arg ...string) {
	wait, done := m.Wait(func(msg *ice.Message, arg ...string) {
		m.Cost(kit.Format("%v->[%v] %v %v", m.Optionv(ice.MSG_SOURCE), name, m.Detailv(), msg.FormatSize())).Copy(msg)
	})
	h := mdb.HashCreate(m.Spawn(), mdb.TYPE, "send", mdb.NAME, kit.Keys(name, m.Target().ID()), mdb.TEXT, kit.Join(arg, ice.SP), kit.Dict(mdb.TARGET, done))
	defer mdb.HashRemove(m, mdb.HASH, h)
	if target := kit.Split(name, ice.PT, ice.PT); mdb.HashSelectDetail(m, target[0], func(value ice.Map) {
		if c, ok := value[mdb.TARGET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotValid, mdb.TARGET) {
			kit.For(m.Optionv(ice.MSG_OPTS), func(k string) { m.Optionv(k, m.Optionv(k)) })
			_space_echo(m.Set(ice.MSG_DETAIL, arg...), []string{h}, target, c)
		}
	}) {
		wait()
	} else {
		m.Warn(kit.IndexOf([]string{ice.OPS, ice.DEV}, target[0]) == -1, ice.ErrNotFound, name)
	}
}

const (
	CHROME = "chrome"
	MASTER = "master"
	SERVER = "server"
	WORKER = "worker"
)
const (
	REDIAL      = "redial"
	SPACE_LOGIN = "space.login"
)
const SPACE = "space"

func init() {
	Index.MergeCommands(ice.Commands{
		SPACE: {Name: "space name cmds auto", Help: "空间站", Actions: ice.MergeActions(ice.Actions{
			tcp.DIAL: {Name: "dial dev=ops name", Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(ice.DEV), HTTP) {
					m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple(ice.DEV))
					m.Option(ice.DEV, ice.DEV)
				}
				_space_dial(m, m.Option(ice.DEV), kit.Select(ice.Info.NodeName, m.Option(mdb.NAME)), arg...)
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				defer mdb.HashModifyDeferRemove(m, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)()
				m.Cmd("", m.Option(mdb.NAME), ice.EXIT)
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					m.Cmds("", func(value ice.Maps) {
						switch value[mdb.TYPE] {
						case MASTER:
							m.PushSearch(mdb.TEXT, m.Cmdv(SPIDE, value[mdb.NAME], CLIENT_ORIGIN), value)
						case SERVER:
							m.PushSearch(mdb.TEXT, MergePods(m, value[mdb.NAME]), value)
						}
					})
				} else if arg[0] == mdb.FOREACH && arg[1] == ssh.SHELL {
					m.PushSearch(mdb.TYPE, ssh.SHELL, mdb.TEXT, "ice.bin space dial dev ops")
				}
			}},
			DOMAIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_space_domain(m)) }},
			LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.MSG_USERUA, m.Cmdv("", kit.Select(m.Option(mdb.NAME), arg, 0), ice.MSG_USERUA))
				m.Cmd("", kit.Select(m.Option(mdb.NAME), arg, 0), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case MASTER:
					ctx.ProcessOpen(m, m.Cmdv(SPIDE, m.Option(mdb.NAME), CLIENT_ORIGIN))
				default:
					ctx.ProcessOpen(m, strings.Split(MergePod(m, m.Option(mdb.NAME), arg), ice.QS)[0])
				}
			}},
			ice.PS: {Hand: func(m *ice.Message, arg ...string) { _space_fork(m) }},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text", ctx.ACTION, OPEN,
			REDIAL, kit.Dict("a", 3000, "b", 1000, "c", 1000),
		), mdb.ClearOnExitHashAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 {
				mdb.HashSelect(m, arg...).Sort("").Table(func(value ice.Maps) {
					m.PushButton(kit.Select(OPEN, LOGIN, value[mdb.TYPE] == LOGIN), mdb.REMOVE)
				})
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
