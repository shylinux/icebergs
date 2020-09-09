package web

import (
	"github.com/gorilla/websocket"
	ice "github.com/shylinux/icebergs"
	aaa "github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/task"

	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"
)

func _space_list(m *ice.Message, space string) {
	if space == "" {
		m.Richs(SPACE, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
			m.PushRender(kit.MDB_LINK, "a", kit.Format(value[kit.MDB_NAME]), kit.MergeURL(m.Option(ice.MSG_USERWEB), "pod", kit.Keys(m.Option("pod"), value[kit.MDB_NAME])))
		})
		m.Sort(kit.MDB_NAME)
		return
	}

	m.Richs(SPACE, nil, space, func(key string, value map[string]interface{}) {
		m.Push("detail", value)
		m.Push(kit.MDB_KEY, kit.MDB_LINK)
		m.PushRender(kit.MDB_VALUE, "a", kit.MergeURL(m.Option(ice.MSG_USERWEB), "pod", kit.Keys(m.Option("pod"), value[kit.MDB_NAME])))
	})
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	m.Richs(SPIDE, nil, dev, func(key string, value map[string]interface{}) {
		client := kit.Value(value, "client").(map[string]interface{})
		redial := m.Confm(SPACE, "meta.redial")
		web := m.Target().Server().(*Frame)

		host := kit.Format(client["hostname"])
		proto := kit.Select("ws", "wss", client["protocol"] == "https")
		uri := kit.MergeURL(proto+"://"+host+"/space/", "name", name, "type", cli.NodeType)
		if u, e := url.Parse(uri); m.Assert(e) {

			task.Put(dev, func(task *task.Task) error {
				for i := 0; i < kit.Int(redial["c"]); i++ {
					if s, e := net.Dial("tcp", host); !m.Warn(e != nil, "%s", e) {
						if s, _, e := websocket.NewClient(s, u, nil, kit.Int(redial["r"]), kit.Int(redial["w"])); !m.Warn(e != nil, "%s", e) {

							// 连接成功
							m.Rich(SPACE, nil, kit.Dict("socket", s,
								kit.MDB_TYPE, MASTER, kit.MDB_NAME, dev, kit.MDB_TEXT, host,
							))
							m.Log_CREATE("space", dev, "retry", i, "uri", uri)

							m = m.Spawns()
							if i = 0; _space_handle(m, true, web.send, s, dev) {
								// 连接关闭
								break
							}
						}
					}

					// 断线重连
					sleep := time.Duration(rand.Intn(kit.Int(redial["a"])*i+2)+kit.Int(redial["b"])) * time.Millisecond
					m.Cost("order: %d sleep: %s reconnect: %s", i, sleep, u)
					time.Sleep(sleep)
				}
				return nil
			})
		}
	})
}
func _space_send(m *ice.Message, space string, arg ...string) {
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
					return m.Cost("[%v]->%v %v %v", id, target, arg, m.Copy(res).Format("append"))
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
	msg.Cost("%v->%v %v %v", source, target, msg.Detailv(), msg.Format("append"))
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
					task.Put(nil, func(task *task.Task) error {
						_space_exec(msg, source, target, c, name)
						return nil
					})
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
			}

			_space_echo(msg, source, target, socket, name)
		}
	}
	return false
}

func _space_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(SPACE, nil, kit.Select(kit.MDB_FOREACH, name), func(key string, value map[string]interface{}) {
		if name != "" && name != value[kit.MDB_NAME] && !strings.Contains(kit.Format(value[kit.MDB_TEXT]), name) {
			return
		}
		m.Push("pod", m.Option("pod"))
		m.Push("ctx", "web")
		m.Push("cmd", SPACE)
		m.Push(key, value, []string{kit.MDB_TIME})
		m.Push(kit.MDB_SIZE, kit.FmtSize(int64(len(kit.Format(value[kit.MDB_TEXT])))))
		m.Push(key, value, []string{kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
	})
}

const (
	MASTER = "master"
	SERVER = "server"
	WORKER = "worker"
	BETTER = "better"
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
			SPACE: {Name: "space [name [cmd...]] auto", Help: "空间站", Action: map[string]*ice.Action{
				"connect": {Name: "connect [dev [name]]", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
					_space_dial(m, kit.Select("dev", arg, 0), kit.Select(cli.NodeName, arg, 2))
				}},
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_space_search(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					_space_list(m, kit.Select("", arg, 0))
					return
				}

				if arg[0] == "" {
					// 本地命令
					m.Cmdy(arg[1:])
					return
				}

				_space_send(m, arg[0], arg[1:]...)
			}},

			"/space/": {Name: "/space/ type name", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if s, e := websocket.Upgrade(m.W, m.R, nil, m.Confi(SPACE, "meta.buffer.r"), m.Confi(SPACE, "meta.buffer.w")); m.Assert(e) {
					name := m.Option(kit.MDB_NAME, strings.Replace(kit.Select(m.Option(ice.MSG_USERADDR), m.Option(kit.MDB_NAME)), ".", "_", -1))
					kind := kit.Select(WORKER, m.Option(kit.MDB_TYPE))

					// 添加节点
					h := m.Rich(SPACE, nil, kit.Dict("socket", s,
						kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, s.RemoteAddr().String(),
					))
					m.Log_CREATE(SPACE, name)

					task.Put(name, func(task *task.Task) error {
						// 监听消息
						switch kind {
						case WORKER:
							m.Event(DREAM_START, name)
							defer m.Event(DREAM_STOP, name)
						default:
							m.Event(gdb.SPACE_START, kind, name)
							defer m.Event(gdb.SPACE_CLOSE, kind, name)
						}
						_space_handle(m, false, m.Target().Server().(*Frame).send, s, name)
						m.Log(ice.LOG_CLOSE, "%s: %s", name, kit.Format(m.Confv(SPACE, kit.Keys(kit.MDB_HASH, h))))
						m.Confv(SPACE, kit.Keys(kit.MDB_HASH, h), "")
						return nil
					})
				}
			}},
		}}, nil)
}
