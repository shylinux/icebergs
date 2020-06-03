package web

import (
	"github.com/gorilla/websocket"
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"
)

var SPACE = ice.Name("space", Index)

func _space_list(m *ice.Message, space string) {
	if space == "" {
		m.Richs(SPACE, nil, "*", func(key string, value map[string]interface{}) {
			m.Push(key, value, []string{"time", "type", "name", "text"})
			if m.Option(ice.MSG_USERUA) != "" {
				m.Push("link", fmt.Sprintf(`<a target="_blank" href="%s?pod=%s">%s</a>`,
					kit.Select(m.Conf(ice.WEB_SHARE, "meta.domain"), m.Option(ice.MSG_USERWEB)),
					kit.Keys(m.Option(ice.MSG_USERPOD), value["name"]), value["name"]))
			}
		})
		m.Sort("name")
		return
	}

	m.Richs(ice.WEB_SPACE, nil, space, func(key string, value map[string]interface{}) {
		m.Push("detail", value)
		m.Push("key", "link")
		m.Push("value", fmt.Sprintf(`<a target="_blank" href="%s?pod=%s">%s</a>`, m.Conf(ice.WEB_SHARE, "meta.domain"), value["name"], value["name"]))
	})
}
func _space_dial(m *ice.Message, dev, name string, arg ...string) {
	// 基本信息
	node := m.Conf(ice.CLI_RUNTIME, "node.type")
	user := m.Conf(ice.CLI_RUNTIME, "boot.username")

	web := m.Target().Server().(*Frame)
	m.Hold(1).Gos(m, func(msg *ice.Message) {
		msg.Richs(ice.WEB_SPIDE, nil, dev, func(key string, value map[string]interface{}) {
			proto := kit.Select("ws", "wss", kit.Format(kit.Value(value, "client.protocol")) == "https")
			host := kit.Format(kit.Value(value, "client.hostname"))

			for i := 0; i < kit.Int(msg.Conf(ice.WEB_SPACE, "meta.redial.c")); i++ {
				if u, e := url.Parse(kit.MergeURL(proto+"://"+host+"/space/", "node", node, "name", name, "user", user, "share", value["share"])); msg.Assert(e) {
					if s, e := net.Dial("tcp", host); !msg.Warn(e != nil, "%s", e) {
						if s, _, e := websocket.NewClient(s, u, nil, kit.Int(msg.Conf(ice.WEB_SPACE, "meta.buffer.r")), kit.Int(msg.Conf(ice.WEB_SPACE, "meta.buffer.w"))); !msg.Warn(e != nil, "%s", e) {
							msg = m.Spawns()

							// 连接成功
							msg.Rich(ice.WEB_SPACE, nil, kit.Dict(
								kit.MDB_TYPE, ice.WEB_MASTER, kit.MDB_NAME, dev, kit.MDB_TEXT, kit.Value(value, "client.hostname"),
								"socket", s,
							))
							msg.Log(ice.LOG_CMDS, "%d conn %s success %s", i, dev, u)
							if i = 0; web.HandleWSS(msg, true, s, dev) {
								break
							}
						}
					}

					// 断线重连
					sleep := time.Duration(rand.Intn(kit.Int(msg.Conf(ice.WEB_SPACE, "meta.redial.a"))*i+2)+kit.Int(msg.Conf(ice.WEB_SPACE, "meta.redial.b"))) * time.Millisecond
					msg.Cost("order: %d sleep: %s reconnect: %s", i, sleep, u)
					time.Sleep(sleep)
				}
			}
		})
		m.Done()
	})
}
func _space_send(m *ice.Message, space string, arg ...string) {
	target := strings.Split(space, ".")
	m.Warn(m.Richs(ice.WEB_SPACE, nil, target[0], func(key string, value map[string]interface{}) {
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
			m.Set(ice.MSG_DETAIL, arg...)
			m.Optionv(ice.MSG_TARGET, target[1:])
			m.Optionv(ice.MSG_SOURCE, []string{id})
			m.Info("send [%s]->%v %v %s", id, target, m.Detailv(), m.Format("meta"))

			// 下发命令
			m.Target().Server().(*Frame).send[id] = m
			socket.WriteMessage(1, []byte(m.Format("meta")))

			m.Option("timeout", m.Conf(SPACE, "meta.timeout.c"))
			m.Call(m.Option("_async") == "", func(res *ice.Message) *ice.Message {
				if res != nil && m != nil {
					// 返回结果
					return m.Copy(res)
				}
				return nil
			})
		}
	}) == nil, "not found %s", space)
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_SPACE: {Name: "space", Help: "空间站", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
				"redial.a", 3000, "redial.b", 1000, "redial.c", 1000,
				"buffer.r", 4096, "buffer.w", 4096,
				"timeout.c", "30s",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_SPACE: {Name: "space name auto", Help: "空间站", Meta: kit.Dict(
				"exports", []string{"pod", "name"},
			), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_space_list(m, "")
					return
				}

				switch arg[0] {
				case "share":
					m.Richs(ice.WEB_SPIDE, nil, m.Option("_dev"), func(key string, value map[string]interface{}) {
						m.Log(ice.LOG_CREATE, "dev: %s share: %s", m.Option("_dev"), arg[1])
						value["share"] = arg[1]
					})

				case "connect":
					_space_dial(m, kit.Select("dev", arg, 1), kit.Select(m.Conf(ice.CLI_RUNTIME, "node.name"), arg, 2))

				default:
					if len(arg) == 1 {
						// 空间详情
						_space_list(m, arg[0])
						break
					}

					if arg[0] == "" {
						// 本地命令
						m.Cmdy(arg[1:])
						break
					}

					_space_send(m, arg[0], arg[1:]...)
				}
			}},

			"/space/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if s, e := websocket.Upgrade(m.W, m.R, nil, m.Confi(ice.WEB_SPACE, "meta.buffer.r"), m.Confi(ice.WEB_SPACE, "meta.buffer.w")); m.Assert(e) {
					m.Option("name", strings.Replace(kit.Select(m.Option(ice.MSG_USERADDR), m.Option("name")), ".", "_", -1))
					m.Option("node", kit.Select("worker", m.Option("node")))

					// 共享空间
					share := m.Option("share")
					if m.Richs(ice.WEB_SHARE, nil, share, nil) == nil {
						share = m.Cmdx(ice.WEB_SHARE, "add", m.Option("node"), m.Option("name"), m.Option("user"))
					}

					// 添加节点
					h := m.Rich(ice.WEB_SPACE, nil, kit.Dict(
						kit.MDB_TYPE, m.Option("node"),
						kit.MDB_NAME, m.Option("name"),
						kit.MDB_TEXT, m.Option("user"),
						"share", share, "socket", s,
					))
					m.Log(ice.LOG_CREATE, "space: %s share: %s", m.Option(kit.MDB_NAME), share)

					m.Gos(m, func(m *ice.Message) {
						// 监听消息
						m.Event(ice.SPACE_START, m.Option("node"), m.Option("name"))
						m.Target().Server().(*Frame).HandleWSS(m, false, s, m.Option("name"))
						m.Log(ice.LOG_CLOSE, "%s: %s", m.Option(kit.MDB_NAME), kit.Format(m.Confv(ice.WEB_SPACE, kit.Keys(kit.MDB_HASH, h))))
						m.Event(ice.SPACE_CLOSE, m.Option("node"), m.Option("name"))
						m.Confv(ice.WEB_SPACE, kit.Keys(kit.MDB_HASH, h), "")
					})

					// 共享空间
					if share != m.Option("share") {
						m.Cmd(ice.WEB_SPACE, m.Option("name"), ice.WEB_SPACE, "share", share)
					}
					m.Echo(share)
				}
			}},
		}}, nil)
}
