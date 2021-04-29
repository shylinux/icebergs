package web

import (
	"github.com/gorilla/websocket"
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"
)

func _space_list(m *ice.Message, space string) {
	m.Fields(space == "", "time,type,name,text")
	m.Cmdy(mdb.SELECT, SPACE, "", mdb.HASH, kit.MDB_NAME, space)

	if space == "" {
		m.Table(func(index int, value map[string]string, head []string) {
			m.PushAnchor(value[kit.MDB_NAME], kit.MergeURL(strings.Split(m.Option(ice.MSG_USERWEB), "?")[0],
				kit.SSH_POD, kit.Keys(m.Option(ice.MSG_USERPOD), value[kit.MDB_NAME])))
		})
		m.SortStrR(kit.MDB_NAME)
	}
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	m.Richs(SPIDE, nil, dev, func(key string, value map[string]interface{}) {
		client := kit.Value(value, tcp.CLIENT).(map[string]interface{})
		redial := m.Confm(SPACE, "meta.redial")
		frame := m.Target().Server().(*Frame)

		host := kit.Format(client[tcp.HOSTNAME])
		proto := strings.Replace(kit.Format(client[tcp.PROTOCOL]), "http", "ws", 1)
		uri := kit.MergeURL(proto+"://"+host+"/space/", kit.MDB_TYPE, ice.Info.NodeType,
			kit.MDB_NAME, name, "share", ice.Info.CtxShare, "river", ice.Info.CtxRiver, arg)

		if u, e := url.Parse(uri); m.Assert(e) {
			m.Go(func() {
				for i := 0; i >= 0 && i < kit.Int(redial["c"]); i++ {
					msg := m.Spawn()
					msg.Option(tcp.DIAL_CB, func(s net.Conn, e error) {
						if msg.Warn(e != nil, e) {
							return
						}

						if s, _, e := websocket.NewClient(s, u, nil, kit.Int(redial["r"]), kit.Int(redial["w"])); !msg.Warn(e != nil, e) {
							msg.Rich(SPACE, nil, kit.Dict(tcp.SOCKET, s, kit.MDB_TYPE, MASTER, kit.MDB_NAME, dev, kit.MDB_TEXT, host))
							msg.Log_CREATE(SPACE, dev, "retry", i, "uri", uri)

							// 连接成功
							if i = 0; _space_handle(msg, true, frame.send, s, dev) {
								i = -2 // 连接关闭
							}
						}
					})
					ls := strings.Split(host, ":")
					msg.Cmd(tcp.CLIENT, tcp.DIAL, kit.MDB_TYPE, "wss", kit.MDB_NAME, dev, tcp.HOST, ls[0], tcp.PORT, ls[1])

					// 断线重连
					sleep := time.Duration(rand.Intn(kit.Int(redial["a"])*i+2)+kit.Int(redial["b"])) * time.Millisecond
					msg.Cost("order", i, "sleep", sleep, "reconnect", u)
					time.Sleep(sleep)
				}
			})
		}
	})
}
func _space_send(m *ice.Message, space string, arg ...string) {
	if space == "" || space == MYSELF || space == ice.Info.NodeName {
		m.Cmdy(arg) // 本地命令
		return
	}

	target := kit.Split(space, ".", ".")
	m.Warn(m.Richs(SPACE, nil, target[0], func(key string, value map[string]interface{}) {
		if socket, ok := value[tcp.SOCKET].(*websocket.Conn); !m.Warn(!ok, ice.ErrNotFound, tcp.SOCKET) {

			// 复制选项
			for _, k := range kit.Simple(m.Optionv(ice.MSG_OPTS)) {
				switch k {
				case ice.MSG_DETAIL, ice.MSG_CMDS, ice.MSG_SESSID:
				default:
					m.Optionv(k, m.Optionv(k))
				}
			}
			m.Optionv(ice.MSG_OPTS, m.Optionv(ice.MSG_OPTS))
			m.Optionv(ice.MSG_OPTION, nil)

			// 构造路由
			frame := m.Target().Server().(*Frame)
			id := kit.Format(m.Target().ID())
			frame.send[id] = m

			// 下发命令
			_space_echo(m.Set(ice.MSG_DETAIL, arg...), []string{id}, target[1:], socket, target[0])

			m.Option("timeout", m.Conf(SPACE, "meta.timeout.c"))
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

func _space_echo(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	msg.Optionv(ice.MSG_SOURCE, source)
	msg.Optionv(ice.MSG_TARGET, target)
	msg.Assert(c.WriteMessage(1, []byte(msg.Format(kit.MDB_META))))

	target = append([]string{name}, target...)
	msg.Log("send", "%v->%v %v %v", source, target, msg.Detailv(), msg.Format(kit.MDB_META))
}
func _space_exec(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	if !msg.Warn(!msg.Right(msg.Detailv()), ice.ErrNotRight) {
		msg = msg.Cmd()
	}

	msg.Set(ice.MSG_OPTS)
	_space_echo(msg, []string{}, kit.Revert(source)[1:], c, name)
	msg.Cost(kit.Format("%v->%v %v %v", source, target, msg.Detailv(), msg.Format(ice.MSG_APPEND)))
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
				if msg.Optionv(ice.MSG_HANDLE, "true"); !msg.Warn(!safe, ice.ErrNotRight) {
					msg.Go(func() { _space_exec(msg, source, target, c, name) })
				}

			} else if msg.Richs(SPACE, nil, target[0], func(key string, value map[string]interface{}) {
				if s, ok := value[tcp.SOCKET].(*websocket.Conn); ok {
					socket, source, target = s, source, target[1:]
					_space_echo(msg, source, target, socket, kit.Select("", target))
					return // 转发报文
				}

				if msg.Warn(msg.Option(ice.MSG_HANDLE) == "true", ice.ErrNotFound) {
					// 回复失败

				} else { // 下发失败
					msg.Warn(true, ice.ErrNotFound)
					source, target = []string{}, kit.Revert(source)[1:]
				}
			}) != nil { // 转发成功

			} else if res, ok := send[msg.Option(ice.MSG_TARGET)]; len(target) != 1 || !ok {
				if msg.Warn(msg.Option(ice.MSG_HANDLE) == "true", ice.ErrNotFound) {
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
func _space_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(SPACE, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		if value = kit.GetMeta(value); strings.Contains(kit.Format(value[kit.MDB_NAME]), name) && value[kit.MDB_TYPE] != MASTER {
			m.PushSearch(kit.SSH_CMD, SPACE, kit.MDB_TYPE, value[kit.MDB_TYPE], kit.MDB_NAME, value[kit.MDB_NAME],
				kit.MDB_TEXT, kit.MergeURL(m.Option(ice.MSG_USERWEB), kit.SSH_POD, kit.Keys(m.Option(ice.MSG_USERPOD), value)))
		}
	})

	m.Cmd(tcp.HOST).Table(func(index int, value map[string]string, head []string) {
		m.PushSearch(kit.SSH_CMD, SPACE, kit.MDB_TYPE, "local", kit.MDB_NAME, value[kit.MDB_NAME],
			kit.MDB_TEXT, "http://"+value[tcp.IP]+":9020", kit.SSH_POD, kit.Keys(m.Option(ice.MSG_USERPOD), value))
	})
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
)
const SPACE = "space"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SPACE: {Name: SPACE, Help: "空间站", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
				"redial", kit.Dict("a", 3000, "b", 1000, "c", 1000, "r", 4096, "w", 4096),
				"timeout", kit.Dict("c", "180s"),
			)},
		},
		Commands: map[string]*ice.Command{
			SPACE: {Name: "space name cmd auto", Help: "空间站", Action: map[string]*ice.Action{
				tcp.DIAL: {Name: "dial dev name", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
					_space_dial(m, m.Option(SPIDE_DEV), kit.Select(ice.Info.NodeName, m.Option(kit.MDB_NAME)))
				}},
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_space_search(m, arg[0], arg[1], kit.Select("", arg, 2))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					_space_list(m, kit.Select("", arg, 0))
					return
				}
				_space_send(m, arg[0], arg[1:]...)
			}},

			"/space/": {Name: "/space/ type name share river", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if s, e := websocket.Upgrade(m.W, m.R, nil, kit.Int(m.Conf(SPACE, "meta.buffer.r")), kit.Int(m.Conf(SPACE, "meta.buffer.w"))); m.Assert(e) {
					name := m.Option(kit.MDB_NAME, strings.Replace(kit.Select(s.RemoteAddr().String(), m.Option(kit.MDB_NAME)), ".", "_", -1))
					kind := kit.Select(WORKER, m.Option(kit.MDB_TYPE))
					share := m.Option("share")
					river := m.Option("river")

					// 添加节点
					h := m.Rich(SPACE, nil, kit.Dict(tcp.SOCKET, s, "share", share, "river", river,
						kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, s.RemoteAddr().String(),
					))

					m.Go(func() {
						defer m.Confv(SPACE, kit.Keys(kit.MDB_HASH, h), "")

						// 监听消息
						switch args := []string{kit.MDB_TYPE, kind, kit.MDB_NAME, name, "share", share, "river", river}; kind {
						case WORKER:
							m.Event(DREAM_START, args...)
							defer m.Event(DREAM_STOP, args...)
						default:
							m.Event(SPACE_START, args...)
							defer m.Event(SPACE_STOP, args...)
						}

						switch kind {
						case CHROME:
							if m.Option(ice.MSG_USERNAME) != "" {
								break
							}
							link := kit.MergeURL(m.Conf(SHARE, kit.Keym(kit.MDB_DOMAIN)), "auth", name)
							go func() {
								m.Sleep("100ms").Cmd(SPACE, name, "pwd", name, link, m.Cmdx(cli.QRCODE, link))
							}()
						}

						frame := m.Target().Server().(*Frame)
						_space_handle(m, false, frame.send, s, name)
					})
				}
			}},
		}})
}
