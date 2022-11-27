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
	msg := m.Cmd(SPIDE, dev)
	host := msg.Append(kit.Keys(tcp.CLIENT, tcp.HOSTNAME))
	proto := strings.Replace(msg.Append(kit.Keys(tcp.CLIENT, tcp.PROTOCOL)), ice.HTTP, "ws", 1)
	uri := kit.ParseURL(kit.MergeURL(proto+"://"+host+PP(SPACE), mdb.TYPE, ice.Info.NodeType, mdb.NAME, name, SHARE, ice.Info.CtxShare, RIVER, ice.Info.CtxRiver, arg))
	m.Go(func() {
		ls := strings.Split(host, ice.DF)
		args := kit.SimpleKV("type,name,host,port", proto, dev, ls[0], kit.Select(kit.Select("443", "80", proto == "ws"), ls, 1))
		redial, _ := m.Configv(REDIAL).(ice.Map)
		a, b, c := kit.Int(redial["a"]), kit.Int(redial["b"]), kit.Int(redial["c"])
		for i := 0; i >= 0 && i < c; i++ {
			msg := m.Spawn()
			msg.Cmd(tcp.CLIENT, tcp.DIAL, args, func(s net.Conn) {
				if s, _, e := websocket.NewClient(s, uri, nil, kit.Int(redial["r"]), kit.Int(redial["w"])); !msg.Warn(e) {
					msg.Logs(mdb.CREATE, SPACE, dev, "retry", i, "uri", uri.String())
					mdb.HashCreate(msg, kit.SimpleKV("", MASTER, dev, host), kit.Dict(mdb.TARGET, s))
					defer mdb.HashRemove(msg, mdb.NAME, name)
					if i = 0; _space_handle(msg, true, m.Target().Server().(*Frame), s, dev) {
						i = -2
					}
				}
			})
			sleep := time.Duration(rand.Intn(a*(i+1))+b) * time.Millisecond
			msg.Cost("order", i, "sleep", sleep, "reconnect", dev)
			if time.Sleep(sleep); mdb.HashSelect(msg).Length() == 0 {
				break
			}
		}
	})
}
func _space_handle(m *ice.Message, safe bool, frame *Frame, c *websocket.Conn, name string) bool {
	for {
		_, b, e := c.ReadMessage()
		if m.Warn(e, SPACE, name) {
			break
		}
		msg := m.Spawn(b)
		socket, source, target := c, kit.Simple(msg.Optionv(ice.MSG_SOURCE), name), kit.Simple(msg.Optionv(ice.MSG_TARGET))
		msg.Log("recv", "%v<-%v %s %v", target, source, msg.Detailv(), msg.FormatMeta())
		if len(target) == 0 {
			if msg.Optionv(ice.MSG_HANDLE, ice.TRUE); safe { // 下行命令
				gdb.Event(msg, SPACE_LOGIN)
			} else { // 上行请求
				msg.Option(ice.MSG_USERROLE, aaa.VOID)
			}
			msg.Go(func() { _space_exec(msg, source, target, c, name) })
		} else if mdb.HashSelectDetail(msg, target[0], func(value ice.Map) {
			if s, ok := value[mdb.TARGET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotValid, target[0]) { // 转发报文
				socket, source, target = s, source, target[1:]
				_space_echo(msg, source, target, socket, kit.Select("", target))
			} else if msg.Option(ice.MSG_HANDLE) != ice.TRUE { // 下发失败
				source, target = []string{}, kit.Revert(source)[1:]
			}
		}) {
		} else if res, ok := frame.getSend(msg.Option(ice.MSG_TARGET)); !m.Warn(!ok || len(target) != 1) {
			back(res, msg) // 接收响应
		}
	}
	return false
}
func _space_domain(m *ice.Message) (link string) {
	if link = ice.Info.Domain; link == "" {
		m.Optionv(ice.MSG_OPTS, ice.MSG_USERNAME)
		link = m.CmdAppend(SPACE, ice.OPS, cli.PWD, mdb.LINK)
	}
	if link == "" {
		link = m.CmdAppend(SPACE, ice.DEV, cli.PWD, mdb.LINK)
	}
	if link == "" {
		link = m.CmdAppend(SPACE, ice.SHY, cli.PWD, mdb.LINK)
	}
	if link == "" {
		link = m.Option(ice.MSG_USERWEB)
	}
	if link == "" {
		link = kit.Format("http://localhost:%s", kit.Select(m.Option(tcp.PORT), m.CmdAppend(SERVE, tcp.PORT)))
	}
	return tcp.PublishLocalhost(m, link)
}
func _space_exec(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	switch msg.Detailv()[0] {
	case "pwd":
		msg.Push(mdb.LINK, kit.MergePOD(_space_domain(msg), name))
	default:
		if aaa.Right(msg, msg.Detailv()) {
			msg = msg.Cmd()
		}
	}
	msg.Set(ice.MSG_OPTS)
	_space_echo(msg, []string{}, kit.Revert(source)[1:], c, name)
	msg.Cost(kit.Format("%v->%v %v %v", source, target, msg.Detailv(), msg.FormatSize()))
}
func _space_echo(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	msg.Optionv(ice.MSG_SOURCE, source)
	msg.Optionv(ice.MSG_TARGET, target)
	if e := c.WriteMessage(1, []byte(msg.FormatMeta())); msg.Warn(e) { // 回复失败
		msg.Go(func() { mdb.HashRemove(msg, mdb.NAME, name) })
		c.Close()
		return
	}
	target = append([]string{name}, target...)
	msg.Log("send", "%v->%v %v %v", source, target, msg.Detailv(), msg.FormatMeta())
}
func _space_send(m *ice.Message, space string, arg ...string) {
	if space == "" || space == MYSELF || space == ice.Info.NodeName {
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
	m.Set(ice.MSG_DETAIL, arg...)
	m.Optionv(ice.MSG_OPTS, m.Optionv(ice.MSG_OPTS))
	m.Optionv(ice.MSG_OPTION, m.Optionv(ice.MSG_OPTS))
	target, id, f := kit.Split(space, ice.PT, ice.PT), "", m.Target().Server().(*Frame)
	if m.Warn(!mdb.HashSelectDetail(m, target[0], func(value ice.Map) {
		if socket, ok := value[mdb.TARGET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotFound, mdb.TARGET) {
			id = f.addSend(kit.Format(m.Target().ID()), m)
			_space_echo(m, []string{id}, target[1:], socket, target[0])
		}
	}), ice.ErrNotFound, space) {
		return
	}
	call(m, m.Config(kit.Keys(TIMEOUT, "c")), func(res *ice.Message) {
		m.Cost(kit.Format("[%v]->%v %v %v", f.delSend(id), target, arg, m.Copy(res).FormatSize()))
	})
}
func _space_fork(m *ice.Message) {
	buffer, _ := m.Configv(BUFFER).(ice.Map)
	if s, e := websocket.Upgrade(m.W, m.R, nil, kit.Int(buffer["r"]), kit.Int(buffer["w"])); m.Assert(e) {
		text := kit.Select(s.RemoteAddr().String(), m.Option(ice.MSG_USERADDR))
		name := strings.ToLower(m.Option(mdb.NAME, kit.ReplaceAll(kit.Select(text, m.Option(mdb.NAME)), ice.PT, "_", ice.DF, "_")))
		kind := kit.Select(WORKER, m.Option(mdb.TYPE))
		args := append([]string{mdb.TYPE, kind, mdb.NAME, name}, m.OptionSimple(SHARE, RIVER, ice.MSG_USERUA)...)
		m.Go(func() {
			mdb.HashCreate(m, mdb.TEXT, kit.Select(text, m.Option(mdb.TEXT)), args, kit.Dict(mdb.TARGET, s))
			defer mdb.HashRemove(m, mdb.NAME, name)
			gdb.Event(m, SPACE_OPEN, args)
			defer gdb.Event(m, SPACE_CLOSE, args)
			switch kind {
			case CHROME:
				m.Go(func(msg *ice.Message) { msg.Sleep30ms(SPACE, name, cli.PWD, name) })
			case WORKER:
				gdb.Event(m, DREAM_START, args)
				defer gdb.Event(m, DREAM_STOP, args)
				if m.Option(cli.DAEMON) == ice.OPS {
					defer m.Cmd(DREAM, DREAM_STOP, args)
				}
			}
			_space_handle(m, false, m.Target().Server().(*Frame), s, name)
		})
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
		PP(SPACE): {Hand: func(m *ice.Message, arg ...string) { _space_fork(m) }},
		SPACE: {Name: "space name cmd auto", Help: "空间站", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Conf("", mdb.HASH, "") }},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { m.Conf("", mdb.HASH, "") }},
			tcp.DIAL: {Name: "dial dev=ops name", Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(ice.DEV), ice.HTTP) {
					m.Cmd(SPIDE, mdb.CREATE, ice.DEV, m.Option(ice.DEV))
					m.Option(ice.DEV, ice.DEV)
				}
				_space_dial(m, m.Option(ice.DEV), kit.Select(ice.Info.NodeName, m.Option(mdb.NAME)), arg...)
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				defer mdb.HashRemove(m, m.OptionSimple(mdb.NAME))
				m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT)
			}},
			SPACE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.MSG_USERROLE, kit.Select(m.Option(ice.MSG_USERROLE), m.CmdAppend(aaa.USER, m.Option(ice.MSG_USERNAME), aaa.USERROLE)))
				if m.Option(ice.MSG_USERROLE) == aaa.VOID && ice.Info.UserName == aaa.VOID {
					m.Option(ice.MSG_USERROLE, aaa.TECH)
				}
				m.Auth(aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
			}},
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case CHROME:
					if m.Option(mdb.NAME) != kit.Select("", kit.Split(m.Option(ice.MSG_DAEMON), ice.PT), 0) {
						m.PushButton(mdb.REMOVE)
					}
				case WORKER, SERVER:
					m.PushButton(OPEN)
				}
			}},
			DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				if arg[1] == m.CommandKey() {
					ProcessWebsite(m, m.Option(mdb.NAME), m.PrefixKey())
				}
			}},
			DOMAIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_space_domain(m)) }},
			OPEN:   {Hand: func(m *ice.Message, arg ...string) { ProcessIframe(m, MergePod(m, m.Option(mdb.NAME)), arg...) }},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text",
			REDIAL, kit.Dict("a", 3000, "b", 1000, "c", 1000), TIMEOUT, kit.Dict("c", "180s"),
			BUFFER, kit.Dict("r", ice.MOD_BUFS, "w", ice.MOD_BUFS),
		), SpaceAction(), DreamAction(), ctx.CmdAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				gdb.Event(m, DREAM_ACTION, arg)
				return
			} else if len(arg) > 1 {
				_space_send(m, strings.ToLower(arg[0]), kit.Simple(kit.Split(arg[1]), arg[2:])...)
				return
			} else if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.Sort("type,name,text")
			}
			if m.IsCliUA() {
				return
			}
			m.Tables(func(value ice.Maps) {
				if msg := gdb.Event(m.Spawn(), DREAM_TABLES, mdb.NAME, value[mdb.NAME], mdb.TYPE, value[mdb.TYPE]); len(msg.Appendv(ctx.ACTION)) > 0 {
					m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
				} else {
					m.PushButton("")
				}
			})
		}},
	})
}

var spaceActions = kit.DictList(SPACE_START, SPACE_OPEN, SPACE_LOGIN, SPACE_CLOSE, SPACE_STOP)

func SpaceAction() ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		for sub := range m.Target().Commands[m.CommandKey()].Actions {
			if spaceActions[sub] == ice.TRUE {
				gdb.Watch(m, sub)
			}
		}
	}}}
}
func Space(m *ice.Message, arg ice.Any) []string {
	if arg == nil || arg == "" || kit.Format(arg) == ice.Info.NodeName {
		return nil
	}
	return []string{SPACE, kit.Format(arg)}
}

func call(m *ice.Message, timeout string, cb func(*ice.Message)) {
	wait := make(chan bool, 2)
	t := time.AfterFunc(kit.Duration(timeout), func() {
		m.Warn(true, ice.ErrNotValid, m.Detailv())
		cb(nil)
		m.Optionv("_cb", nil)
		wait <- false
	})
	m.Optionv("_cb", func(res *ice.Message) {
		cb(res)
		t.Stop()
		wait <- true
	})
	<-wait
}
func back(m *ice.Message, res *ice.Message) {
	switch cb := m.Optionv("_cb").(type) {
	case func(*ice.Message):
		cb(res)
	}
}
