package web

import (
	"math/rand"
	"net"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/websocket"
)

func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	msg := m.Cmd(SPIDE, tcp.CLIENT, dev, PP(SPACE))
	uri := kit.ParseURL(strings.Replace(kit.MergeURL(msg.Append(DOMAIN), mdb.TYPE, ice.Info.NodeType, mdb.NAME, name, SHARE, ice.Info.CtxShare, RIVER, ice.Info.CtxRiver), ice.HTTP, "ws", 1))
	args := kit.SimpleKV("type,name,host,port", msg.Append(tcp.PROTOCOL), dev, msg.Append(tcp.HOST), msg.Append(tcp.PORT))
	m.Go(func() {
		redial := kit.Dict(m.Configv(REDIAL))
		a, b, c := kit.Int(redial["a"]), kit.Int(redial["b"]), kit.Int(redial["c"])
		for i := 1; i < c; i++ {
			next := time.Duration(rand.Intn(a*(i+1))+b*i) * time.Millisecond
			m.Cmd(tcp.CLIENT, tcp.DIAL, args, func(c net.Conn) {
				if conn, _, e := websocket.NewClient(c, uri, nil, kit.Int(redial["r"]), kit.Int(redial["w"])); !m.Warn(e, tcp.DIAL, dev, SPACE, uri.String()) {
					defer mdb.HashCreateDeferRemove(m, kit.SimpleKV("", MASTER, dev, msg.Append(tcp.HOSTNAME)), kit.Dict(mdb.TARGET, conn))()
					_space_handle(m.Spawn(), true, dev, conn)
					i = 0
				}
			}).Cost("order", i, "sleep", next, "redial", dev, "uri", uri.String()).Sleep(next)
		}
	})
}
func _space_fork(m *ice.Message) {
	buffer := kit.Dict(m.Configv(BUFFER))
	if conn, e := websocket.Upgrade(m.W, m.R, nil, kit.Int(buffer["r"]), kit.Int(buffer["w"])); m.Assert(e) {
		text := kit.Select(m.Option(ice.MSG_USERADDR), m.Option(mdb.TEXT))
		name := strings.ToLower(kit.ReplaceAll(kit.Select(m.Option(ice.MSG_USERADDR), m.Option(mdb.NAME)), ice.PT, "_", ice.DF, "_"))
		args := kit.Simple(mdb.TYPE, kit.Select(WORKER, m.Option(mdb.TYPE)), mdb.NAME, name, mdb.TEXT, text, m.OptionSimple(SHARE, RIVER, ice.MSG_USERUA))
		m.Go(func() {
			defer mdb.HashCreateDeferRemove(m, args, kit.Dict(mdb.TARGET, conn))()
			defer gdb.EventDeferEvent(m, SPACE_OPEN, args)(SPACE_CLOSE, args)
			switch m.Option(mdb.TYPE) {
			case WORKER:
				defer gdb.EventDeferEvent(m, DREAM_OPEN, args)(DREAM_CLOSE, args)
			case CHROME:
				m.Go(func() { m.Sleep30ms().Cmd(SPACE, name, cli.PWD, name) })
			}
			_space_handle(m, false, name, conn)
		})
	}
}
func _space_handle(m *ice.Message, safe bool, name string, conn *websocket.Conn) {
	for {
		_, b, e := conn.ReadMessage()
		if e != nil {
			m.Cost(SPACE, name, e)
			break
		}
		msg := m.Spawn(b)
		source, target := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name), kit.Simple(msg.Optionv(ice.MSG_TARGET))
		msg.Log("recv", "%v->%v %v %v", source, target, msg.Detailv(), msg.FormatMeta())
		if next := msg.Option(ice.MSG_TARGET); next == "" || len(target) == 0 {
			if msg.Optionv(ice.MSG_HANDLE, ice.TRUE); safe { // 下行命令
				gdb.Event(msg, SPACE_LOGIN)
			} else { // 上行请求
				msg.Option(ice.MSG_USERROLE, aaa.VOID)
			}
			msg.Go(func() { _space_exec(msg, source, target, conn) })
		} else if mdb.HashSelectDetail(msg, next, func(value ice.Map) {
			if conn, ok := value[mdb.TARGET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotValid, next) {
				_space_echo(msg, source, target, conn) // 转发报文
			}
		}) {
		} else if res := getSend(m, next); !m.Warn(res == nil || len(target) != 1, ice.ErrNotFound, next) {
			res.Cost(kit.Format("[%v]->%v %v %v", next, res.Optionv(ice.MSG_TARGET), res.Detailv(), msg.FormatSize()))
			back(res, msg.Sleep("10ms")) // 接收响应
		}
	}
}
func _space_domain(m *ice.Message) (link string) {
	return kit.GetValid(
		func() string { return ice.Info.Domain },
		func() string {
			return m.CmdAppend(SPACE, ice.OPS, cli.PWD, kit.Dict(ice.MSG_OPTS, ice.MSG_USERNAME), mdb.LINK)
		},
		func() string { return m.CmdAppend(SPACE, ice.DEV, cli.PWD, mdb.LINK) },
		func() string { return m.CmdAppend(SPACE, ice.SHY, cli.PWD, mdb.LINK) },
		func() string { return tcp.PublishLocalhost(m, m.Option(ice.MSG_USERWEB)) },
		func() string {
			return kit.Format("http://%s:%s", m.CmdAppend(tcp.HOST, aaa.IP), kit.Select(m.Option(tcp.PORT), m.CmdAppend(SERVE, tcp.PORT)))
		})
}
func _space_exec(msg *ice.Message, source, target []string, conn *websocket.Conn) {
	switch kit.Select(cli.PWD, msg.Detailv(), 0) {
	case cli.PWD:
		msg.Push(mdb.LINK, kit.MergePOD(_space_domain(msg), kit.Select("", source, -1)))
	default:
		if aaa.Right(msg, msg.Detailv()) {
			msg = msg.Cmd()
		}
	}
	defer msg.Cost(kit.Format("%v->%v %v %v", source, target, msg.Detailv(), msg.FormatSize()))
	_space_echo(msg.Set(ice.MSG_OPTS), []string{}, kit.Revert(kit.Simple(source)), conn)
}
func _space_echo(m *ice.Message, source, target []string, conn *websocket.Conn) {
	if m.Options(ice.MSG_SOURCE, source, ice.MSG_TARGET, target[1:]); m.Warn(conn.WriteMessage(1, []byte(m.FormatMeta()))) {
		mdb.HashRemove(m, mdb.NAME, target[0])
	} else {
		m.Log("send", "%v->%v %v %v", source, target, m.Detailv(), m.FormatMeta())
	}
}
func _space_send(m *ice.Message, space string, arg ...string) {
	if space == "" || space == ice.Info.NodeName {
		m.Cmdy(arg)
		return
	}
	kit.Simple(m.Optionv(ice.MSG_OPTS), func(k string) {
		switch k {
		case ice.MSG_DETAIL, ice.MSG_CMDS, ice.MSG_SESSID:
		default:
			m.Optionv(k, m.Optionv(k))
		}
	})
	m.Set(ice.MSG_DETAIL, arg...).Optionv(ice.MSG_OPTION, m.Optionv(ice.MSG_OPTS, m.Optionv(ice.MSG_OPTS)))
	target := kit.Split(space, ice.PT, ice.PT)
	if mdb.HashSelectDetail(m, target[0], func(value ice.Map) {
		if conn, ok := value[mdb.TARGET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotValid, mdb.TARGET) {
			_space_echo(m, []string{addSend(m, m)}, target, conn)
		}
	}) {
		call(m, m.Config(kit.Keys(TIMEOUT, "c")), func(res *ice.Message) { m.Copy(res) })
	} else if kit.IndexOf([]string{ice.OPS, ice.DEV, ice.SHY}, target[0]) > -1 {
		return
	} else {
		m.Warn(true, ice.ErrNotFound, space)
	}
}

const (
	CHROME = "chrome"
	FRIEND = "friend"
	MASTER = "master"
	MYSELF = "myself"
	SERVER = "server"
	WORKER = "worker"
)
const (
	BUFFER  = "buffer"
	REDIAL  = "redial"
	TIMEOUT = "timeout"

	SPACE_START = "space.start"
	SPACE_OPEN  = "space.open"
	SPACE_LOGIN = "space.login"
	SPACE_CLOSE = "space.close"
	SPACE_STOP  = "space.stop"
)
const SPACE = "space"

func init() {
	Index.MergeCommands(ice.Commands{
		SPACE: {Name: "space name cmd auto", Help: "空间站", Actions: ice.MergeActions(ice.Actions{
			tcp.DIAL: {Name: "dial dev=ops name", Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(ice.DEV), ice.HTTP) {
					m.Cmd(SPIDE, mdb.CREATE, ice.DEV, m.Option(ice.DEV))
					m.Option(ice.DEV, ice.DEV)
				}
				_space_dial(m, m.Option(ice.DEV), kit.Select(ice.Info.NodeName, m.Option(mdb.NAME)), arg...)
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				defer mdb.HashModifyDeferRemove(m, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)()
				m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT)
			}},
			SPACE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				if ice.Info.Username == aaa.VOID {
					m.Option(ice.MSG_USERROLE, aaa.TECH)
				} else {
					m.Option(ice.MSG_USERROLE, kit.Select(m.Option(ice.MSG_USERROLE), m.CmdAppend(aaa.USER, m.Option(ice.MSG_USERNAME), aaa.USERROLE)))
				}
				aaa.SessAuth(m, kit.Dict(aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERROLE, m.Option(ice.MSG_USERROLE)))
			}},
			DOMAIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_space_domain(m)) }},
			OPEN:   {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessOpen(m, MergePod(m, m.Option(mdb.NAME), arg)) }},
			ice.PS: {Hand: func(m *ice.Message, arg ...string) { _space_fork(m) }},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text", ctx.ACTION, OPEN,
			REDIAL, kit.Dict("a", 3000, "b", 1000, "c", 1000), TIMEOUT, kit.Dict("c", "10s"),
			BUFFER, kit.Dict("r", ice.MOD_BUFS, "w", ice.MOD_BUFS),
		), mdb.ClearHashOnExitAction(), SpaceAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 {
				mdb.HashSelect(m, arg...).Sort("type,name,text")
			} else {
				_space_send(m, strings.ToLower(arg[0]), kit.Simple(kit.Split(arg[1]), arg[2:])...)
			}
		}},
	})
}

func SpaceAction() ice.Actions {
	return gdb.EventsAction(SPACE_START, SPACE_OPEN, SPACE_LOGIN, SPACE_CLOSE, SPACE_STOP)
}
func Space(m *ice.Message, arg ice.Any) []string {
	if arg == nil || arg == "" || kit.Format(arg) == ice.Info.NodeName {
		return nil
	}
	return []string{SPACE, kit.Format(arg)}
}
func call(m *ice.Message, timeout string, cb func(*ice.Message)) bool {
	wait := make(chan bool, 2)
	t := time.AfterFunc(kit.Duration(timeout), func() {
		m.Warn(true, ice.ErrNotValid, m.Detailv())
		m.Optionv("_cb", nil)
		cb(nil)
		wait <- false
	})
	m.Optionv("_cb", func(res *ice.Message) {
		t.Stop()
		cb(res)
		wait <- true
	})
	return <-wait
}
func back(m *ice.Message, res *ice.Message) bool {
	switch cb := m.Optionv("_cb").(type) {
	case func(*ice.Message):
		cb(res)
		return true
	}
	return false
}
func addSend(m *ice.Message, msg *ice.Message) string {
	return m.Target().Server().(*Frame).addSend(kit.Format(m.Target().ID()), msg)
}
func getSend(m *ice.Message, key string) *ice.Message {
	return m.Target().Server().(*Frame).getSend(key)
}
