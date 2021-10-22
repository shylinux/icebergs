package web

import (
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _space_link(m *ice.Message, pod string, arg ...interface{}) string {
	return tcp.ReplaceLocalhost(m, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/chat/pod/"+pod, arg...))
}
func _space_domain(m *ice.Message) (link string) {
	link = m.Config(kit.MDB_DOMAIN)
	if link == "" {
		link = m.Cmd(SPACE, ice.DEV, cli.PWD).Append(kit.MDB_LINK)
	}
	if link == "" {
		link = m.Cmd(SPACE, ice.SHY, cli.PWD).Append(kit.MDB_LINK)
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
	m.Richs(SPIDE, nil, dev, func(key string, value map[string]interface{}) {
		client := kit.Value(value, tcp.CLIENT).(map[string]interface{})
		redial := m.Confm(SPACE, kit.Keym("redial"))
		frame := m.Target().Server().(*Frame)

		host := kit.Format(client[tcp.HOSTNAME])
		proto := strings.Replace(kit.Format(client[tcp.PROTOCOL]), "http", "ws", 1)
		uri := kit.MergeURL(proto+"://"+host+"/space/", kit.MDB_TYPE, ice.Info.NodeType,
			kit.MDB_NAME, name, SHARE, ice.Info.CtxShare, RIVER, kit.Select(ice.Info.CtxRiver, m.Option(RIVER)), arg)
		u := kit.ParseURL(uri)

		m.Go(func() {
			for i := 0; i >= 0 && i < kit.Int(redial["c"]); i++ {
				msg := m.Spawn()
				msg.Option(kit.Keycb(tcp.DIAL), func(s net.Conn, e error) {
					if msg.Warn(e != nil, e) {
						return
					}

					if s, _, e := websocket.NewClient(s, u, nil, kit.Int(redial["r"]), kit.Int(redial["w"])); !msg.Warn(e != nil, e) {
						msg.Rich(SPACE, nil, kit.Dict(SOCKET, s, kit.MDB_TYPE, MASTER, kit.MDB_NAME, dev, kit.MDB_TEXT, host))
						msg.Log_CREATE(SPACE, dev, "retry", i, "uri", uri)

						// 连接成功
						if i = 0; _space_handle(msg, true, frame.send, s, dev) {
							i = -2 // 连接关闭
						}
					}
				})
				ls := strings.Split(host, ":")
				msg.Cmd(tcp.CLIENT, tcp.DIAL, kit.MDB_TYPE, "wss", kit.MDB_NAME, dev, tcp.HOST, ls[0], tcp.PORT, kit.Select("443", ls, 1))

				// 断线重连
				sleep := time.Duration(rand.Intn(kit.Int(redial["a"])*i+2)+kit.Int(redial["b"])) * time.Millisecond
				msg.Cost("order", i, "sleep", sleep, "reconnect", u)
				time.Sleep(sleep)
			}
		})
	})
}
func _space_handle(m *ice.Message, safe bool, send map[string]*ice.Message, c *websocket.Conn, name string) bool {
	for running := true; running; {
		if _, b, e := c.ReadMessage(); m.Warn(e != nil, e) {
			break
		} else {
			socket, msg := c, m.Spawn(b)
			target := kit.Simple(msg.Optionv(ice.MSG_TARGET))
			source := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name)
			msg.Log("recv", "%v->%v %s %v", source, target, msg.Detailv(), msg.Format(kit.MDB_META))

			if len(target) == 0 { // 本地执行
				msg.Log_AUTH(aaa.USERROLE, msg.Option(ice.MSG_USERROLE), aaa.USERNAME, msg.Option(ice.MSG_USERNAME))
				if msg.Optionv(ice.MSG_HANDLE, ice.TRUE); safe {
					msg.Go(func() { _space_exec(msg, source, target, c, name) })
				} else {
					msg.Push(kit.MDB_LINK, kit.MergePOD(_space_domain(msg), name))
					_space_echo(msg, []string{}, kit.Revert(source)[1:], c, name)
				}

			} else if msg.Richs(SPACE, nil, target[0], func(key string, value map[string]interface{}) {
				if s, ok := value[SOCKET].(*websocket.Conn); ok {
					socket, source, target = s, source, target[1:]
					_space_echo(msg, source, target, socket, kit.Select("", target))
					return // 转发报文
				}

				if msg.Warn(msg.Option(ice.MSG_HANDLE) == ice.TRUE, ice.ErrNotFound) {
					// 回复失败

				} else { // 下发失败
					msg.Warn(true, ice.ErrNotFound)
					source, target = []string{}, kit.Revert(source)[1:]
				}
			}) != nil { // 转发成功

			} else if res, ok := send[msg.Option(ice.MSG_TARGET)]; len(target) != 1 || !ok {
				if msg.Warn(msg.Option(ice.MSG_HANDLE) == ice.TRUE, ice.ErrNotFound) {
					// 回复失败

				} else { // 下发失败
					msg.Warn(true, ice.ErrNotFound)
					source, target = []string{}, kit.Revert(source)[1:]
				}

			} else { // 接收响应
				m.Sleep("30ms")
				res.Back(msg)
			}
		}
	}
	return false
}
func _space_exec(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	if !msg.Warn(!msg.Right(msg.Detailv()), ice.ErrNotRight) {
		msg = msg.Cmd()
	}

	msg.Set(ice.MSG_OPTS)
	_space_echo(msg, []string{}, kit.Revert(source)[1:], c, name)
	msg.Cost(kit.Format("%v->%v %v %v", source, target, msg.Detailv(), msg.Format(ice.MSG_APPEND)))
}
func _space_echo(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	msg.Optionv(ice.MSG_SOURCE, source)
	msg.Optionv(ice.MSG_TARGET, target)
	msg.Assert(c.WriteMessage(1, []byte(msg.Format(kit.MDB_META))))

	target = append([]string{name}, target...)
	msg.Log("send", "%v->%v %v %v", source, target, msg.Detailv(), msg.Format(kit.MDB_META))
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

			m.Option("timeout", m.Config("timeout.c"))
			m.Call(m.Option("_async") == "", func(res *ice.Message) *ice.Message {
				// 返回结果
				if delete(frame.send, id); res != nil && m != nil {
					return m.Cost(kit.Format("[%v]->%v %v %v", id, target, arg, m.Copy(res).Format(ice.MSG_APPEND)))
				}
				return nil
			})
		}
	}) == nil, ice.ErrNotFound, space)
}
func _space_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(SPACE, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		if value = kit.GetMeta(value); !strings.Contains(kit.Format(value[kit.MDB_NAME]), name) {
			return
		}

		switch value[kit.MDB_TYPE] {
		case CHROME:

		case MASTER:
			m.PushSearch(ice.CMD, SPACE, kit.MDB_TYPE, value[kit.MDB_TYPE], kit.MDB_NAME, value[kit.MDB_NAME],
				kit.MDB_TEXT, m.Cmd(SPIDE, value[kit.MDB_NAME], ice.OptionFields("client.url")).Append("client.url"), value)

		default:
			m.PushSearch(ice.CMD, SPACE, kit.MDB_TYPE, value[kit.MDB_TYPE], kit.MDB_NAME, value[kit.MDB_NAME],
				kit.MDB_TEXT, _space_link(m, kit.Keys(m.Option(ice.MSG_USERPOD), value[kit.MDB_NAME])), value)
		}
	})

	if name == "" {
		port := m.Cmd(SERVE, ice.Option{mdb.FIELDS, tcp.PORT}).Append(tcp.PORT)
		if port == "" {
			return
		}
		m.Cmd(tcp.HOST).Table(func(index int, value map[string]string, head []string) {
			m.PushSearch(ice.CMD, SPACE, kit.MDB_TYPE, MYSELF, kit.MDB_NAME, value[kit.MDB_NAME],
				kit.MDB_TEXT, kit.Format("http://%s:%s", value[aaa.IP], port))
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

	SOCKET = "socket"
)
const SPACE = "space"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SPACE: {Name: SPACE, Help: "空间站", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,type,name,text",
			"redial", kit.Dict("a", 3000, "b", 1000, "c", 1000, "r", ice.MOD_BUFS, "w", ice.MOD_BUFS),
			"timeout", kit.Dict("c", "180s"),
		)},
	}, Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf(SPACE, kit.MDB_HASH, "")
			m.Cmd(mdb.SEARCH, mdb.CREATE, SPACE, m.Prefix(SPACE))
		}},
		SPACE: {Name: "space name cmd auto", Help: "空间站", Action: ice.MergeAction(map[string]*ice.Action{
			tcp.DIAL: {Name: "dial dev name river", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				_space_dial(m, m.Option(ice.DEV), kit.Select(ice.Info.NodeName, m.Option(kit.MDB_NAME)))
			}},
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_space_search(m, arg[0], arg[1], kit.Select("", arg, 2))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 { // 节点列表
				if mdb.HashSelect(m, arg...); len(arg) == 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushAnchor(value[kit.MDB_NAME], _space_link(m, kit.Keys(m.Option(ice.MSG_USERPOD), value[kit.MDB_NAME])))
					})
					m.SortStrR(kit.MDB_NAME)
				}
				return
			}
			// 下发命令
			_space_send(m, arg[0], arg[1:]...)
		}},
		"/space/": {Name: "/space/ type name share river", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if s, e := websocket.Upgrade(m.W, m.R, nil, kit.Int(m.Config("buffer.r")), kit.Int(m.Config("buffer.w"))); m.Assert(e) {
				name := kit.Select(s.RemoteAddr().String(), m.Option(kit.MDB_NAME))
				name = m.Option(kit.MDB_NAME, strings.Replace(name, ".", "_", -1))
				name = m.Option(kit.MDB_NAME, strings.Replace(name, ":", "-", -1))
				kind := kit.Select(WORKER, m.Option(kit.MDB_TYPE))

				// 添加节点
				args := append([]string{kit.MDB_TYPE, kind, kit.MDB_NAME, name}, m.OptionSimple(SHARE, RIVER)...)
				h := m.Rich(SPACE, nil, kit.Dict(SOCKET, s, kit.MDB_TEXT, s.RemoteAddr().String(), args))

				m.Go(func() {
					defer m.Confv(SPACE, kit.Keys(kit.MDB_HASH, h), "")

					switch kind {
					case CHROME: // 交互节点
						m.Go(func(msg *ice.Message) {
							switch m.Option("cmd") {
							case "pwd":
								link := kit.MergeURL(_space_domain(msg), "grant", name)
								msg.Sleep("100ms").Cmd(SPACE, name, "pwd", name, link, msg.Cmdx(cli.QRCODE, link))
							default:
								msg.Sleep("100ms").Cmd(SPACE, name, "pwd", name)
							}
						})
					case WORKER: // 工作节点
						m.Event(DREAM_START, args...)
						defer m.Event(DREAM_STOP, args...)
					default: // 服务节点
						m.Event(SPACE_START, args...)
						defer m.Event(SPACE_STOP, args...)
					}

					frame := c.Server().(*Frame)
					_space_handle(m, false, frame.send, s, name)
				})
			}
		}},
	}})
}
