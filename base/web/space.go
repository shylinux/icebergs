package web

import (
	"github.com/gorilla/websocket"
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"math/rand"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

func _space_list(m *ice.Message, space string) {
	m.Option(mdb.FIELDS, kit.Select("time,type,name,text", mdb.DETAIL, space != ""))
	m.Cmdy(mdb.SELECT, SPACE, "", mdb.HASH, kit.MDB_HASH, space)
	m.Table(func(index int, value map[string]string, head []string) {
		if p := kit.MergeURL(m.Option(ice.MSG_USERWEB), kit.SSH_POD, kit.Keys(m.Option(kit.SSH_POD), kit.Select(value[kit.MDB_VALUE], value[kit.MDB_NAME]))); space == "" {
			m.PushRender(kit.MDB_LINK, "a", value[kit.MDB_NAME], p)
		} else if value[kit.MDB_KEY] == kit.MDB_NAME {
			m.Push(kit.MDB_KEY, kit.MDB_LINK)
			m.PushRender(kit.MDB_VALUE, "a", value[kit.MDB_VALUE], p)
		}
	})
	m.Sort(kit.MDB_NAME)
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	m.Debug("what %v %v %v", dev, name, arg)
	m.Richs(SPIDE, nil, dev, func(key string, value map[string]interface{}) {
		m.Debug("what")
		client := kit.Value(value, "client").(map[string]interface{})
		redial := m.Confm(SPACE, "meta.redial")
		web := m.Target().Server().(*Frame)

		host := kit.Format(client["hostname"])
		proto := kit.Select("ws", "wss", client["protocol"] == "https")
		uri := kit.MergeURL(proto+"://"+host+"/space/", "name", name, "type", ice.Info.NodeType, "share", os.Getenv("ctx_share"), "river", os.Getenv("ctx_river"))
		if u, e := url.Parse(uri); m.Assert(e) {
			m.Debug("what")

			m.Go(func() {
				m.Debug("what")
				for i := 0; i >= 0 && i < kit.Int(redial["c"]); i++ {
					m.Option(tcp.DIAL_CB, func(s net.Conn, e error) {
						if m.Warn(e != nil, e) {
							return
						}
						if s, _, e := websocket.NewClient(s, u, nil, kit.Int(redial["r"]), kit.Int(redial["w"])); !m.Warn(e != nil, e) {
							m.Rich(SPACE, nil, kit.Dict("socket", s, kit.MDB_TYPE, MASTER, kit.MDB_NAME, dev, kit.MDB_TEXT, host))
							m.Log_CREATE(SPACE, dev, "retry", i, "uri", uri)

							// 连接成功
							m = m.Spawns()
							if i = 0; _space_handle(m, true, web.send, s, dev) {
								i = -1 // 连接关闭
							}
						}
					})
					ls := strings.Split(host, ":")
					m.Cmd(tcp.CLIENT, tcp.DIAL, kit.MDB_TYPE, "wss", kit.MDB_NAME, dev, tcp.HOST, ls[0], tcp.PORT, ls[1])

					// 断线重连
					sleep := time.Duration(rand.Intn(kit.Int(redial["a"])*i+2)+kit.Int(redial["b"])) * time.Millisecond
					m.Cost("order", i, "sleep", sleep, "reconnect", u)
					time.Sleep(sleep)
				}
			})
		}
	})
}
func _space_send(m *ice.Message, space string, arg ...string) {
	if space == "" || space == MYSELF || space == ice.Info.NodeName {
		m.Cmdy(arg)
		return // 本地命令
	}

	target := strings.Split(space, ".")
	frame := m.Target().Server().(*Frame)
	m.Warn(m.Richs(SPACE, nil, target[0], func(key string, value map[string]interface{}) {
		if socket, ok := value["socket"].(*websocket.Conn); !m.Warn(!ok, "socket err") {
			// 复制选项
			for _, k := range kit.Simple(m.Optionv("_option")) {
				switch k {
				case "detail", "cmds", ice.MSG_SESSID:
				default:
					m.Optionv(k, m.Optionv(k))
				}
			}
			m.Optionv("_option", m.Optionv("_option"))
			m.Optionv("option", nil)

			// 构造路由
			id := kit.Format(m.Target().ID())
			frame.send[id] = m

			// 下发命令
			m.Set(ice.MSG_DETAIL, arg...)
			_space_echo(m, []string{id}, target[1:], socket, target[0])

			m.Option("timeout", m.Conf(SPACE, "meta.timeout.c"))
			m.Call(m.Option("_async") == "", func(res *ice.Message) *ice.Message {
				if delete(frame.send, id); res != nil && m != nil {
					// 返回结果
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
	e := c.WriteMessage(1, []byte(msg.Format("meta")))
	msg.Assert(e)
	target = append([]string{name}, target...)
	msg.Log("send", "%v->%v %v %v", source, target, msg.Detailv(), msg.Format("meta"))
}
func _space_exec(msg *ice.Message, source, target []string, c *websocket.Conn, name string) {
	if !msg.Warn(!msg.Right(msg.Detailv()), ice.ErrNotAuth) {
		msg = msg.Cmd()
	}
	msg.Set("_option")
	_space_echo(msg, []string{}, kit.Revert(source)[1:], c, name)
	msg.Cost(kit.Format("%v->%v %v %v", source, target, msg.Detailv(), msg.Format(ice.MSG_APPEND)))
}
func _space_handle(m *ice.Message, safe bool, send map[string]*ice.Message, c *websocket.Conn, name string) bool {
	for running := true; running; {
		if t, b, e := c.ReadMessage(); m.Warn(e != nil, "space recv %d msg %v", t, e) {
			// 解析失败
			break
		} else {
			socket, msg := c, m.Spawns(b)
			target := kit.Simple(msg.Optionv(ice.MSG_TARGET))
			source := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name)
			msg.Log("recv", "%v<-%v %s %v", target, source, msg.Detailv(), msg.Format("meta"))

			if len(target) == 0 {
				if msg.Option(ice.MSG_USERROLE, aaa.UserRole(msg, msg.Option(ice.MSG_USERNAME))) == aaa.VOID {
					role := msg.Cmdx(SPIDE, SPIDE_DEV, SPIDE_MSG, SPIDE_GET, "/chat/header", "cmds", aaa.USERROLE, "who", msg.Option(ice.MSG_USERNAME))
					msg.Option(ice.MSG_USERROLE, kit.Select(role, aaa.TECH, role == aaa.ROOT))
				}
				msg.Log_AUTH(aaa.USERROLE, msg.Option(ice.MSG_USERROLE), aaa.USERNAME, msg.Option(ice.MSG_USERNAME))

				if msg.Optionv(ice.MSG_HANDLE, "true"); !msg.Warn(!safe, ice.ErrNotAuth) {
					// 本地执行
					msg.Option("_dev", name)
					msg.Go(func() { _space_exec(msg, source, target, c, name) })
					continue
				}
			} else if msg.Richs(SPACE, nil, target[0], func(key string, value map[string]interface{}) {
				// 查询节点
				if s, ok := value["socket"].(*websocket.Conn); ok {
					socket, source, target = s, source, target[1:]
				} else {
					socket, source, target = s, source, target[1:]
				}
			}) != nil {
				// 转发报文

			} else if res, ok := send[msg.Option(ice.MSG_TARGET)]; len(target) == 1 && ok {
				// 接收响应
				res.Back(msg)
				continue

			} else if msg.Warn(msg.Option("_handle") == "true", "space miss") {
				// 回复失败
				continue

			} else {
				// 下发失败
				msg.Warn(true, "space error")
				source, target = []string{}, kit.Revert(source)[1:]
				continue
			}

			_space_echo(msg, source, target, socket, name)
		}
	}
	return false
}

const (
	MASTER = "master"
	MYSELF = "myself"
	BETTER = "better"
	CHROME = "chrome"
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
				"connect": {Name: "connect dev name", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
					_space_dial(m, arg[0], kit.Select(ice.Info.NodeName, arg, 1))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					_space_list(m, kit.Select("", arg, 0))
					return
				}
				_space_send(m, arg[0], arg[1:]...)
			}},

			"/space/": {Name: "/space/ type name", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if s, e := websocket.Upgrade(m.W, m.R, nil, kit.Int(m.Conf(SPACE, "meta.buffer.r")), kit.Int(m.Conf(SPACE, "meta.buffer.w"))); m.Assert(e) {
					name := m.Option(kit.MDB_NAME, strings.Replace(kit.Select(m.Option(ice.MSG_USERADDR), m.Option(kit.MDB_NAME)), ".", "_", -1))
					kind := kit.Select(WORKER, m.Option(kit.MDB_TYPE))
					share := m.Option("share")
					river := m.Option("river")

					// 添加节点
					h := m.Rich(SPACE, nil, kit.Dict("socket", s, "share", share, "river", river,
						kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, s.RemoteAddr().String(),
					))

					m.Go(func() {
						// 监听消息
						switch kind {
						case WORKER:
							m.Event(DREAM_START, "type", kind, "name", name, "share", share, "river", river)
							defer m.Event(DREAM_STOP, "type", kind, "name", name, "share", share, "river", river)
						default:
							m.Event(SPACE_START, "type", kind, "name", name, "share", share, "river", river)
							defer m.Event(SPACE_STOP, "type", kind, "name", name, "share", share, "river", river)
						}

						frame := m.Target().Server().(*Frame)
						_space_handle(m, false, frame.send, s, name)
						m.Confv(SPACE, kit.Keys(kit.MDB_HASH, h), "")
					})
				}
			}},
		}}, nil)
}
