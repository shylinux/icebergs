package web

import (
	"math/rand"
	"net"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/websocket"
)

func _space_link(m *ice.Message, pod string, arg ...interface{}) string {
	return tcp.ReplaceLocalhost(m, kit.MergePOD(m.Option(ice.MSG_USERWEB), pod, arg...))
}
func _space_domain(m *ice.Message) (link string) {
	link = m.Config(DOMAIN)
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
		link = kit.Format("http://localhost:%s", m.Cmd(SERVE).Append(tcp.PORT))
	}
	return tcp.ReplaceLocalhost(m, link)
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	if strings.HasPrefix(dev, "http") {
		m.Cmd(SPIDE, mdb.CREATE, ice.DEV, dev)
		dev = ice.DEV
	}

	value := m.Richs(SPIDE, nil, dev, nil)
	client := kit.Value(value, tcp.CLIENT).(map[string]interface{})

	host := kit.Format(client[tcp.HOSTNAME])
	proto := strings.Replace(kit.Format(client[tcp.PROTOCOL]), "http", "ws", 1)
	uri := kit.MergeURL(proto+"://"+host+"/space/", mdb.TYPE, ice.Info.NodeType, mdb.NAME, name,
		SHARE, m.Conf(cli.RUNTIME, kit.Keys("conf.ctx_share")), RIVER, m.Conf(cli.RUNTIME, kit.Keys("conf.ctx_river")), arg)

	m.Go(func() {
		u := kit.ParseURL(uri)
		redial := m.Configm(REDIAL)
		frame := m.Target().Server().(*Frame)

		for i := 0; i >= 0 && i < kit.Int(redial["c"]); i++ {
			msg := m.Spawn()
			ls := strings.Split(host, ":")
			msg.Cmd(tcp.CLIENT, tcp.DIAL, kit.SimpleKV("type,name,host,port", proto, dev, ls[0], kit.Select("443", ls, 1)), func(s net.Conn) {
				if s, _, e := websocket.NewClient(s, u, nil, kit.Int(redial["r"]), kit.Int(redial["w"])); !msg.Warn(e) {
					msg.Rich(SPACE, nil, kit.Dict(SOCKET, s, kit.SimpleKV("", MASTER, dev, host)))
					msg.Log_CREATE(SPACE, dev, "retry", i, "uri", uri)

					// 连接成功
					if i = 0; _space_handle(msg, true, frame.send, s, dev) {
						i = -2 // 连接关闭
					}
				}
			})

			// 断线重连
			sleep := time.Duration(rand.Intn(kit.Int(redial["a"])*i+2)+kit.Int(redial["b"])) * time.Millisecond
			msg.Cost("order", i, "sleep", sleep, "reconnect", dev)
			time.Sleep(sleep)
		}
	})
}
func _space_handle(m *ice.Message, safe bool, send map[string]*ice.Message, c *websocket.Conn, name string) bool {
	for running := true; running; {
		if _, b, e := c.ReadMessage(); m.Warn(e, SPACE, name) {
			break
		} else {
			socket, msg := c, m.Spawn(b)
			target := kit.Simple(msg.Optionv(ice.MSG_TARGET))
			source := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name)
			msg.Log("recv", "%v->%v %s %v", source, target, msg.Detailv(), msg.FormatMeta())

			if len(target) == 0 {
				msg.Log_AUTH(aaa.USERROLE, msg.Option(ice.MSG_USERROLE), aaa.USERNAME, msg.Option(ice.MSG_USERNAME))
				if msg.Optionv(ice.MSG_HANDLE, ice.TRUE); safe { // 下行命令
					msg.Go(func() { _space_exec(msg, source, target, c, name) })
				} else { // 上行请求
					msg.Push(mdb.LINK, kit.MergePOD(_space_domain(msg), name))
					_space_echo(msg, []string{}, kit.Revert(source)[1:], c, name)
				}

			} else if msg.Richs(SPACE, nil, target[0], func(key string, value map[string]interface{}) {
				if s, ok := value[SOCKET].(*websocket.Conn); ok {
					socket, source, target = s, source, target[1:]
					_space_echo(msg, source, target, socket, kit.Select("", target))
					return // 转发报文
				}

				if msg.Warn(msg.Option(ice.MSG_HANDLE) == ice.TRUE, ice.ErrNotFound, "already handled") {
					// 回复失败

				} else { // 下发失败
					msg.Warn(true, ice.ErrNotFound, target)
					source, target = []string{}, kit.Revert(source)[1:]
				}
			}) != nil { // 转发成功

			} else if res, ok := send[msg.Option(ice.MSG_TARGET)]; len(target) != 1 || !ok {
				if msg.Warn(msg.Option(ice.MSG_HANDLE) == ice.TRUE, ice.ErrNotFound, target) {
					// 回复失败

				} else { // 下发失败
					msg.Warn(true, ice.ErrNotFound, target)
					source, target = []string{}, kit.Revert(source)[1:]
				}

			} else { // 接收响应
				m.Sleep30ms()
				res.Back(msg)
			}
		}
	}
	return false
}
func _space_exec(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	if msg.Right(msg.Detailv()) {
		msg = msg.Cmd()
	}

	msg.Set(ice.MSG_OPTS)
	_space_echo(msg, []string{}, kit.Revert(source)[1:], c, name)
	msg.Cost(kit.Format("%v->%v %v %v", source, target, msg.Detailv(), msg.FormatSize()))
}
func _space_echo(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	msg.Optionv(ice.MSG_SOURCE, source)
	msg.Optionv(ice.MSG_TARGET, target)
	msg.Assert(c.WriteMessage(1, []byte(msg.FormatMeta())))

	target = append([]string{name}, target...)
	msg.Log("send", "%v->%v %v %v", source, target, msg.Detailv(), msg.FormatMeta())
}
func _space_send(m *ice.Message, space string, arg ...string) {
	if space == "" || space == MYSELF || space == ice.Info.NodeName {
		m.Cmdy(arg) // 本地命令
		return
	}

	target := kit.Split(space, ice.PT, ice.PT)
	m.Warn(m.Richs(SPACE, nil, target[0], func(key string, value map[string]interface{}) {
		if socket, ok := value[SOCKET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotFound, SOCKET) {

			// 复制选项
			for _, k := range kit.Simple(m.Optionv(ice.MSG_OPTS)) {
				switch k {
				case ice.MSG_DETAIL, ice.MSG_CMDS, ice.MSG_SESSID:
				default:
					m.Optionv(k, m.Optionv(k))
				}
			}
			m.Optionv(ice.MSG_OPTS, m.Optionv(ice.MSG_OPTS))
			m.Optionv(ice.MSG_OPTION, m.Optionv(ice.MSG_OPTS))

			// 构造路由
			frame := m.Target().Server().(*Frame)
			id := kit.Format(m.Target().ID())
			frame.send[id] = m

			// 下发命令
			_space_echo(m.Set(ice.MSG_DETAIL, arg...), []string{id}, target[1:], socket, target[0])

			m.Option(TIMEOUT, m.Config(kit.Keys(TIMEOUT, "c")))
			m.Call(m.Option("_async") == "", func(res *ice.Message) *ice.Message { // 返回结果
				if delete(frame.send, id); res != nil && m != nil {
					return m.Cost(kit.Format("[%v]->%v %v %v", id, target, arg, m.Copy(res).FormatSize()))
				}
				return nil
			})
		}
	}) == nil, ice.ErrNotFound, space)
}
func _space_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(SPACE, nil, mdb.FOREACH, func(key string, value map[string]interface{}) {
		if value = kit.GetMeta(value); !strings.Contains(kit.Format(value[mdb.NAME]), name) {
			return
		}

		switch value[mdb.TYPE] {
		case CHROME:

		case MASTER:
			m.PushSearch(mdb.TEXT, m.Cmd(SPIDE, value[mdb.NAME], ice.OptionFields("")).Append("client.url"), value)
		default:
			m.PushSearch(mdb.TEXT, _space_link(m, kit.Keys(m.Option(ice.MSG_USERPOD), value[mdb.NAME])), value)
		}
	})
	if name != "" {
		return
	}
	m.Cmd(SERVE, ice.OptionFields("")).Table(func(index int, val map[string]string, head []string) {
		m.Cmd(tcp.HOST, ice.OptionFields("")).Table(func(index int, value map[string]string, head []string) {
			m.PushSearch(kit.SimpleKV("", MYSELF, value[mdb.NAME], kit.Format("http://%s:%s", value[aaa.IP], val[tcp.PORT])))
		})
	})
}
func _space_fork(m *ice.Message) {
	if s, e := websocket.Upgrade(m.W, m.R, nil, kit.Int(m.Config("buffer.r")), kit.Int(m.Config("buffer.w"))); m.Assert(e) {
		text := kit.Select(s.RemoteAddr().String(), m.Option(ice.MSG_USERADDR))
		name := m.Option(mdb.NAME, kit.ReplaceAll(kit.Select(text, m.Option(mdb.NAME)), ".", "_", ":", "_"))
		kind := kit.Select(WORKER, m.Option(mdb.TYPE))
		args := append([]string{mdb.TYPE, kind, mdb.NAME, name}, m.OptionSimple(SHARE, RIVER)...)

		m.Go(func() {
			h := m.Rich(SPACE, nil, kit.Dict(SOCKET, s, mdb.TEXT, text, args))
			m.Log_CREATE(SPACE, name, mdb.TYPE, kind)

			switch kind {
			case CHROME: // 交互节点
				defer m.Confv(SPACE, kit.Keys(mdb.HASH, h), "")
				m.Go(func(msg *ice.Message) {
					switch m.Option(ice.CMD) {
					case cli.PWD:
						link := kit.MergeURL(_space_domain(msg), aaa.GRANT, name)
						msg.Sleep300ms(SPACE, name, cli.PWD, name, link, msg.Cmdx(cli.QRCODE, link))
					case "sso":
						link := _space_domain(msg)
						ls := strings.Split(kit.ParseURL(link).Path, ice.PS)
						link = kit.MergeURL2(_space_domain(msg), "/chat/sso", "space", kit.Select("", ls, 3), "back", m.Option(ice.MSG_USERWEB))
						msg.Sleep300ms(SPACE, name, cli.PWD, name, link, msg.Cmdx(cli.QRCODE, link))
					default:
						msg.Sleep300ms(SPACE, name, cli.PWD, name)
					}
				})
			case WORKER: // 工作节点
				m.Event(DREAM_START, args...)
				defer m.Event(DREAM_STOP, args...)
				defer m.Cmd(DREAM, DREAM_STOP, args)
			default: // 服务节点
				m.Event(SPACE_START, args...)
				defer m.Event(SPACE_STOP, args...)
			}

			_space_handle(m, false, m.Target().Server().(*Frame).send, s, name)
		})
	}
}

const (
	CHROME = "chrome"
	MASTER = "master"
	MYSELF = "myself"
	SERVER = "server"
	WORKER = "worker"
)
const (
	SPACE_START = "space.start"
	SPACE_STOP  = "space.stop"

	SOCKET  = "socket"
	BUFFER  = "buffer"
	REDIAL  = "redial"
	TIMEOUT = "timeout"
)
const SPACE = "space"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SPACE: {Name: SPACE, Help: "空间站", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text",
			BUFFER, kit.Dict("r", ice.MOD_BUFS, "w", ice.MOD_BUFS),
			REDIAL, kit.Dict("a", 3000, "b", 1000, "c", 1000), TIMEOUT, kit.Dict("c", "180s"),
		)},
	}, Commands: map[string]*ice.Command{
		SPACE: {Name: "space name cmd auto invite", Help: "空间站", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, m.CommandKey(), m.PrefixKey())
				m.Conf(SPACE, mdb.HASH, "")
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
				_space_dial(m, m.Option(ice.DEV), kit.Select(ice.Info.NodeName, m.Option(mdb.NAME)))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 { // 节点列表
				if mdb.HashSelect(m, arg...); len(arg) == 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushAnchor(value[mdb.NAME], _space_link(m, kit.Keys(m.Option(ice.MSG_USERPOD), value[mdb.NAME])))
					})
					m.SortStrR(mdb.NAME)
				}
				return
			}
			// 下发命令
			_space_send(m, arg[0], arg[1:]...)
		}},
		"/space/": {Name: "/space/ type name share river", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_space_fork(m)
		}},
	}})
}
