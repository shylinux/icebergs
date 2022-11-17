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
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/websocket"
)

func _space_domain(m *ice.Message) (link string) {
	if link = ice.Info.Domain; link == "" {
		m.Optionv(ice.MSG_OPTS, ice.MSG_USERNAME)
		link = m.Cmd(SPACE, ice.OPS, cli.PWD).Append(mdb.LINK)
	}
	if link == "" {
		link = m.Cmd(SPACE, ice.DEV, cli.PWD).Append(mdb.LINK)
	}
	if link == "" {
		link = m.Cmd(SPACE, ice.SHY, cli.PWD).Append(mdb.LINK)
	}
	if link == "" {
		link = m.Option(ice.MSG_USERWEB)
	}
	if link == "" {
		link = kit.Format("http://localhost:%s", kit.Select(m.Option(tcp.PORT), m.Cmd(SERVE).Append(tcp.PORT)))
	}
	return tcp.ReplaceLocalhost(m, link)
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	if strings.HasPrefix(dev, ice.HTTP) {
		m.Cmd(SPIDE, mdb.CREATE, ice.DEV, dev)
		dev = ice.DEV
	}

	msg := m.Cmd(SPIDE, dev)
	host := msg.Append(kit.Keys(tcp.CLIENT, tcp.HOSTNAME))
	proto := strings.Replace(msg.Append(kit.Keys(tcp.CLIENT, tcp.PROTOCOL)), ice.HTTP, "ws", 1)
	uri := kit.MergeURL(proto+"://"+host+PP(SPACE), mdb.TYPE, ice.Info.NodeType, mdb.NAME, name, SHARE, ice.Info.CtxShare, RIVER, ice.Info.CtxRiver, arg)
	u := kit.ParseURL(uri)

	m.Go(func() {
		frame := m.Target().Server().(*Frame)

		ls := strings.Split(host, ice.DF)
		args := kit.SimpleKV("type,name,host,port", proto, dev, ls[0], kit.Select("443", ls, 1))

		redial, _ := m.Configv(REDIAL).(ice.Map)
		a, b, c := kit.Int(redial["a"]), kit.Int(redial["b"]), kit.Int(redial["c"])
		for i := 0; i >= 0 && i < c; i++ {
			msg := m.Spawn()
			msg.Cmd(tcp.CLIENT, tcp.DIAL, args, func(s net.Conn) {
				if s, _, e := websocket.NewClient(s, u, nil, kit.Int(redial["r"]), kit.Int(redial["w"])); !msg.Warn(e) {
					msg.Logs(mdb.CREATE, SPACE, dev, "retry", i, "uri", uri)
					mdb.HashCreate(msg, kit.SimpleKV("", MASTER, dev, host), kit.Dict(mdb.TARGET, s))
					defer mdb.HashRemove(msg, mdb.NAME, name)

					if i = 0; _space_handle(msg, true, frame, s, dev) {
						i = -2 // 关闭连接
					}
				}
			})

			// 断线重连
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

		socket, msg := c, m.Spawn(b)
		target := kit.Simple(msg.Optionv(ice.MSG_TARGET))
		source := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name)
		msg.Log("recv", "%v<-%v %s %v", target, source, msg.Detailv(), msg.FormatMeta())

		if len(target) == 0 { // 执行命令
			if msg.Optionv(ice.MSG_HANDLE, ice.TRUE); safe { // 下行命令
				msg.Option(ice.MSG_USERROLE, kit.Select(msg.Option(ice.MSG_USERROLE), msg.Cmd(aaa.USER, msg.Option(ice.MSG_USERNAME)).Append(aaa.USERROLE)))
				if msg.Option(ice.MSG_USERROLE) == aaa.VOID && ice.Info.UserName == aaa.TECH {
					msg.Option(ice.MSG_USERROLE, aaa.TECH) // 演示空间
				}
				msg.Auth(aaa.USERROLE, msg.Option(ice.MSG_USERROLE), aaa.USERNAME, msg.Option(ice.MSG_USERNAME))
				msg.Go(func() { _space_exec(msg, source, target, c, name) })
				continue
			}
			// 上行请求
			msg.Push(mdb.LINK, kit.MergePOD(_space_domain(msg), name))
			_space_echo(msg, []string{}, kit.Revert(source)[1:], c, name)
			continue
		}

		if mdb.HashSelectDetail(msg, target[0], func(value ice.Map) { // 转发命令
			if s, ok := value[mdb.TARGET].(*websocket.Conn); ok {
				socket, source, target = s, source, target[1:]
				_space_echo(msg, source, target, socket, kit.Select("", target))
				return // 转发报文
			}

			if msg.Warn(msg.Option(ice.MSG_HANDLE) == ice.TRUE, ice.ErrNotValid, "already handled") {
				// 回复失败

			} else { // 下发失败
				msg.Warn(true, ice.ErrNotFound, target)
				source, target = []string{}, kit.Revert(source)[1:]
			}
		}) {
			continue
		}

		if res, ok := frame.getSend(msg.Option(ice.MSG_TARGET)); len(target) != 1 || !ok {
			if msg.Warn(msg.Option(ice.MSG_HANDLE) == ice.TRUE, ice.ErrNotValid, target) {
				// 回复失败

			} else { // 下发失败
				msg.Warn(true, ice.ErrNotFound, target)
				source, target = []string{}, kit.Revert(source)[1:]
			}
			continue
		} else { // 接收响应
			m.Sleep30ms()
			back(res, msg)
		}
	}
	return false
}
func _space_exec(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	if aaa.Right(msg, msg.Detailv()) { // 执行命令
		msg = msg.Cmd()
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
		m.Cmdy(arg) // 本地命令
		return
	}

	// 生成参数
	for _, k := range kit.Simple(m.Optionv(ice.MSG_OPTS)) {
		switch k {
		case ice.MSG_DETAIL, ice.MSG_CMDS, ice.MSG_SESSID:
		default:
			m.Optionv(k, m.Optionv(k))
		}
	}
	m.Optionv(ice.MSG_OPTS, m.Optionv(ice.MSG_OPTS))
	m.Optionv(ice.MSG_OPTION, m.Optionv(ice.MSG_OPTS))
	m.Set(ice.MSG_DETAIL, arg...)

	// 发送命令
	frame := m.Target().Server().(*Frame)
	target, id := kit.Split(space, ice.PT, ice.PT), ""
	if m.Warn(!mdb.HashSelectDetail(m, target[0], func(value ice.Map) {
		if socket, ok := value[mdb.TARGET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotFound, mdb.TARGET) {
			id = frame.addSend(kit.Format(m.Target().ID()), m)
			_space_echo(m, []string{id}, target[1:], socket, target[0])
		}
	}), ice.ErrNotFound, space) {
		return
	}

	// 返回结果
	m.Option(TIMEOUT, m.Config(kit.Keys(TIMEOUT, "c")))
	call(m, m.Option("_async") == "", func(res *ice.Message) {
		m.Cost(kit.Format("[%v]->%v %v %v", id, target, arg, m.Copy(res).FormatSize()))
		frame.delSend(id)
	})
}
func _space_fork(m *ice.Message) {
	buffer, _ := m.Configv(BUFFER).(ice.Map)
	if s, e := websocket.Upgrade(m.W, m.R, nil, kit.Int(buffer["r"]), kit.Int(buffer["w"])); m.Assert(e) {
		text := kit.Select(s.RemoteAddr().String(), m.Option(ice.MSG_USERADDR))
		name := strings.ToLower(m.Option(mdb.NAME, kit.ReplaceAll(kit.Select(text, m.Option(mdb.NAME)), ".", "_", ":", "_")))
		kind := kit.Select(WORKER, m.Option(mdb.TYPE))
		args := append([]string{mdb.TYPE, kind, mdb.NAME, name}, m.OptionSimple(SHARE, RIVER, ice.MSG_USERUA)...)

		m.Go(func() {
			mdb.HashCreate(m, mdb.TEXT, kit.Select(text, m.Option(mdb.TEXT)), args, kit.Dict(mdb.TARGET, s))
			defer mdb.HashRemove(m, mdb.NAME, name)

			gdb.Event(m, SPACE_OPEN, args)
			defer gdb.Event(m, SPACE_CLOSE, args)

			switch kind {
			case CHROME:
				m.Go(func(msg *ice.Message) { msg.Sleep300ms(SPACE, name, cli.PWD, name) })
			case WORKER:
				gdb.Event(m, DREAM_START, args)
				defer gdb.Event(m, DREAM_STOP, args)
				if m.Option("daemon") == "ops" {
					defer m.Cmd(DREAM, DREAM_STOP, args)
				}
			}
			_space_handle(m, false, m.Target().Server().(*Frame), s, name)
		})
	}
}
func _space_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Cmd(SPACE, ice.OptionFields(""), func(value ice.Maps) {
		if !strings.Contains(value[mdb.NAME], name) {
			return
		}
		switch value[mdb.TYPE] {
		case CHROME:
		case MASTER:
			m.PushSearch(mdb.TEXT, m.CmdAppend(SPIDE, value[mdb.NAME], CLIENT_URL), value)
		default:
			m.PushSearch(mdb.TEXT, MergePod(m, value[mdb.NAME]), value)
		}
	})
	if name != "" {
		return
	}
	m.Cmd(SERVE, ice.OptionFields(""), func(val ice.Maps) {
		m.Cmd(tcp.HOST, ice.OptionFields(""), func(value ice.Maps) {
			m.PushSearch(kit.SimpleKV("", MYSELF, value[mdb.NAME], kit.Format("http://%s:%s", value[aaa.IP], val[tcp.PORT])))
		})
	})
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

	SPACE_OPEN  = "space.open"
	SPACE_CLOSE = "space.close"
	SPACE_START = "space.start"
	SPACE_STOP  = "space.stop"
)
const SPACE = "space"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		SPACE: {Name: SPACE, Help: "空间站", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text",
			BUFFER, kit.Dict("r", ice.MOD_BUFS, "w", ice.MOD_BUFS),
			REDIAL, kit.Dict("a", 3000, "b", 1000, "c", 1000), TIMEOUT, kit.Dict("c", "180s"),
		)},
	}, Commands: ice.Commands{
		SPACE: {Name: "space name cmd auto invite", Help: "空间站", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Conf("", mdb.HASH, "") }},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectClose(m)
				m.Conf("", mdb.HASH, "")
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				defer mdb.HashRemove(m, m.OptionSimple(mdb.NAME))
				m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT)
			}},
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_space_search(m, arg[0], arg[1], kit.Select("", arg, 2))
			}},
			aaa.INVITE: {Name: "invite", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				for _, k := range []string{ice.MISC, ice.CORE, ice.BASE} {
					m.Cmdy("web.code.publish", ice.CONTEXTS, k)
				}

				m.EchoScript("shell", "# 共享环境", m.Option(ice.MSG_USERWEB))
				m.EchoAnchor(m.Option(ice.MSG_USERWEB)).Echo(ice.NL)
				m.EchoQRCode(m.Option(ice.MSG_USERWEB))
			}},
			tcp.DIAL: {Name: "dial dev=ops name", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				_space_dial(m, m.Option(ice.DEV), kit.Select(ice.Info.NodeName, m.Option(mdb.NAME)), arg...)
			}},
			DOMAIN: {Name: "domain", Help: "域名", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_space_domain(m))
			}},
			cli.OPEN: {Name: "open", Help: "系统", Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, MergePod(m, m.Option(mdb.NAME), "", ""), arg...)
			}},
			"xterm": {Name: "xterm", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, MergePodCmd(m, m.Option(mdb.NAME), "web.code.xterm", mdb.HASH,
					m.Cmdx(SPACE, m.Option(mdb.NAME), "web.code.xterm", mdb.CREATE, mdb.TYPE, nfs.SH, m.OptionSimple(mdb.NAME), mdb.TEXT, "")), arg...)
			}},
			"vimer": {Name: "vimer", Help: "编程", Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, MergePodCmd(m, m.Option(mdb.NAME), "web.code.vimer", "", ""), arg...)
			}},
			"exit": {Name: "exit", Help: "关闭", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", m.Option(mdb.NAME), "close")
				m.Cmd("", m.Option(mdb.NAME), "exit")
				ctx.ProcessRefresh(m)
			}},
		}, mdb.HashCloseAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 { // 节点列表
				if mdb.HashSelect(m, arg...); len(arg) == 0 {
					m.Tables(func(value ice.Maps) {
						switch value[mdb.TYPE] {
						case SERVER, WORKER:
							m.PushButton(cli.OPEN, "xterm", "vimer")
						case CHROME:
							if value[mdb.NAME] == kit.Select("", kit.Split(m.Option(ice.MSG_DAEMON), ice.PT), 0) {
								m.PushButton("")
							} else {
								m.PushButton("exit")
							}
						default:
							m.PushButton("")
						}
					})
					m.Sort("type,name,text")
				}
				return
			}
			// 下发命令
			_space_send(m, strings.ToLower(arg[0]), arg[1:]...)
		}},
		PP(SPACE): {Name: "/space/ type name share river", Help: "空间站", Hand: func(m *ice.Message, arg ...string) {
			_space_fork(m)
		}},
	}})
}
func Space(m *ice.Message, arg ice.Any) []string {
	if arg == nil || arg == "" || kit.Format(arg) == ice.Info.NodeName {
		return nil
	}
	return []string{SPACE, kit.Format(arg)}
}

func call(m *ice.Message, sync bool, cb func(*ice.Message)) {
	wait := make(chan bool, 2)

	p := kit.Select("10s", m.Option(TIMEOUT))
	t := time.AfterFunc(kit.Duration(p), func() {
		m.Warn(true, ice.ErrNotValid, m.Detailv())
		back(m, nil)
		wait <- false
	})

	m.Optionv("_cb", func(res *ice.Message) {
		if cb(res); sync {
			wait <- true
			t.Stop()
		}
	})

	if sync {
		<-wait
	} else {
		t.Stop()
	}
}
func back(m *ice.Message, res *ice.Message) {
	switch cb := m.Optionv("_cb").(type) {
	case func(*ice.Message):
		cb(res)
	}
}
