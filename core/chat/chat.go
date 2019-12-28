package chat

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "chat", Help: "聊天模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.CHAT_RIVER: {Name: "river", Help: "群组", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", "chat.json")

			if m.Conf(ice.CLI_RUNTIME, "boot.count") == "1" {
				// 系统群组
				m.Option(ice.MSG_USERNAME, m.Conf(ice.CLI_RUNTIME, "boot.username"))
				m.Option(ice.MSG_USERROLE, ice.ROLE_ROOT)
				river := m.Cmdx("web.chat./ocean", "spawn", "meet", m.Conf(ice.CLI_RUNTIME, "boot.username"))
				river = m.Cmdx("web.chat./steam", river, "spawn", "miss",
					"", "", "spide", "",
					"", "", "space", "",
					"", "", "dream", "",
					"", "", "favor", "",
					"", "", "story", "",
					"", "", "share", "",
				)

				// 用户权限
				m.Cmd(ice.AAA_ROLE, "white", ice.ROLE_VOID, "enable", "/river")
				m.Cmd(ice.AAA_ROLE, "white", ice.ROLE_VOID, "enable", "/storm")
				m.Cmd(ice.AAA_ROLE, "black", ice.ROLE_VOID, "enable", "/storm", "rename")
				m.Cmd(ice.AAA_ROLE, "black", ice.ROLE_VOID, "enable", "/storm", "remove")

				m.Cmd(ice.AAA_ROLE, "white", ice.ROLE_VOID, "enable", "/action")
				m.Cmd(ice.AAA_ROLE, "white", ice.ROLE_VOID, "enable", "dream")
				m.Cmd(ice.AAA_ROLE, "black", ice.ROLE_VOID, "enable", "dream.停止")
			}

			// 用户群组
			m.Watch(ice.USER_CREATE, "web.chat./ocean", "spawn", "")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "chat.json", ice.CHAT_RIVER)
		}},

		ice.WEB_LOGIN: {Name: "login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				switch arg[0] {
				case "login":
					// 用户登录
					m.Option(ice.MSG_SESSID, web.Cookie(m, m.Cmdx(ice.AAA_USER, "login", m.Option(ice.MSG_USERNAME, arg[1]), arg[2])))
				default:
					// 用户群组
					m.Richs(ice.CHAT_RIVER, nil, arg[0], func(value map[string]interface{}) {
						if m.Option(ice.MSG_RIVER, arg[0]); len(arg) > 1 {
							m.Richs(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "tool"), arg[1], func(value map[string]interface{}) {
								m.Option(ice.MSG_STORM, arg[1])
							})
						}
						m.Info("river: %s storm: %s", m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM))
					})
				}
			}
			if m.Option(ice.MSG_USERURL) == "/login" {
				return
			}

			// 登录检查
			if m.Warn(!m.Options(ice.MSG_SESSID) || !m.Options(ice.MSG_USERNAME), "not login") {
				m.Option(ice.MSG_USERURL, "")
			}
			// 权限检查
			if !m.Right(m.Option(ice.MSG_USERURL), m.Optionv("cmds")) {
				m.Option(ice.MSG_USERURL, "")
			}
		}},

		"/toast": {Name: "/toast", Help: "提示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/tutor": {Name: "/tutor", Help: "向导", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/debug": {Name: "/debug", Help: "调试", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/carte": {Name: "/carte", Help: "菜单", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/favor": {Name: "/favor", Help: "收藏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Hand = false
			if msg := m.Cmd(arg); !msg.Hand {
				m.Set("result").Cmdy(ice.CLI_SYSTEM, arg)
			} else {
				m.Copy(msg)
			}
		}},
		"/login": {Name: "/login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "check":
				m.Echo(m.Option(ice.MSG_USERNAME))
			case "login":
				m.Echo(m.Option(ice.MSG_SESSID))
			}
		}},

		"/ocean": {Name: "/ocean", Help: "大海洋", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 用户列表
				m.Richs(ice.AAA_USER, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"username", "usernode"})
				})
				return
			}

			switch arg[0] {
			case "spawn":
				// 创建群组
				river := m.Rich(ice.CHAT_RIVER, nil, kit.Dict(
					kit.MDB_META, kit.Dict(kit.MDB_NAME, kit.Select(arg[2]+"@"+m.Conf(ice.CLI_RUNTIME, "boot.hostname"), arg[1])),
					"user", kit.Data(kit.MDB_SHORT, "username"),
					"tool", kit.Data(),
				))
				m.Info("create river: %v name: %v", river, arg[1])
				m.Cmd("/river", river, "add", arg[2:])
				m.Echo(river)
			}
		}},
		"/river": {Name: "/river", Help: "小河流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch len(arg) {
			case 0:
				// 群组列表
				m.Richs(ice.CHAT_RIVER, nil, "*", func(key string, value map[string]interface{}) {
					m.Richs(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, key, "user"), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
						m.Push(key, value["meta"], []string{kit.MDB_KEY, kit.MDB_NAME})
					})
				})
			case 1:
				// 群组详情
				m.Richs(ice.CHAT_RIVER, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"], []string{kit.MDB_KEY, kit.MDB_NAME})
				})
			default:
				switch arg[1] {
				case "add":
					m.Rich(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "user"), kit.Data("username", m.Conf(ice.CLI_RUNTIME, "boot.username")))
					// 添加用户
					for _, v := range arg[2:] {
						user := m.Rich(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "user"), kit.Data("username", v))
						m.Info("add river: %s user: %s name: %s", arg[0], user, v)
					}
				}
			}
		}},
		"/storm": {Name: "/storm", Help: "暴风雨", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Keys(kit.MDB_HASH, arg[0], "tool")
			if len(arg) < 2 {
				// 应用列表
				m.Richs(ice.CHAT_RIVER, prefix, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"], []string{kit.MDB_KEY, kit.MDB_NAME})
				})
				m.Sort(kit.MDB_NAME)
				return
			}

			if !m.Right(cmd, arg[2]) {
				return
			}
			switch arg[2] {
			case "add":
				// 添加命令
				for i := 3; i < len(arg)-3; i += 4 {
					id := m.Grow(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1]), kit.Data(
						"pod", arg[i], "ctx", arg[i+1], "cmd", arg[i+2], "key", arg[i+3],
					))
					m.Info("create tool %d %v", id, arg[i:i+4])
				}
			case "rename":
				// 重命名应用
				old := m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1], kit.MDB_META, kit.MDB_NAME))
				m.Info("rename storm: %s %s->%s", arg[1], old, arg[3])
				m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1], kit.MDB_META, kit.MDB_NAME), arg[3])

			case "remove":
				// 删除应用
				m.Richs(ice.CHAT_RIVER, kit.Keys(prefix), arg[1], func(value map[string]interface{}) {
					m.Info("remove storm: %s %s", arg[1], kit.Format(value))
				})
				m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1]), "")
			}
		}},
		"/steam": {Name: "/steam", Help: "大气层", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				// 设备列表
				m.Richs(ice.WEB_SPACE, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"type", "name", "user"})
				})
				return
			}

			switch arg[1] {
			case "spawn":
				// 创建应用
				storm := m.Rich(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "tool"), kit.Dict(
					kit.MDB_META, kit.Dict(kit.MDB_NAME, arg[2]),
				))
				m.Info("create river: %s storm: %s name: %v", arg[0], storm, arg[2])
				m.Cmd("/storm", arg[0], storm, "add", arg[3:])
				m.Echo(storm)

			default:
				// 命令列表
				m.Cmdy(ice.WEB_SPACE, arg[2], ice.CTX_COMMAND)
			}
		}},

		"/action": {Name: "/action", Help: "小工具", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Keys(kit.MDB_HASH, arg[0], "tool", kit.MDB_HASH, arg[1])
			if len(arg) == 2 {
				// 命令列表
				m.Set(ice.MSG_OPTION)
				m.Grows(ice.CHAT_RIVER, prefix, "", "", func(index int, value map[string]interface{}) {
					if meta, ok := kit.Value(value, "meta").(map[string]interface{}); ok {
						m.Push("river", arg[0])
						m.Push("storm", arg[1])
						m.Push("action", index)

						m.Push("node", meta["pod"])
						m.Push("group", meta["ctx"])
						m.Push("index", meta["cmd"])

						msg := m.Cmd(m.Space(meta["pod"]), ice.CTX_COMMAND, meta["ctx"], meta["cmd"])
						m.Push("name", meta["cmd"])
						m.Push("help", msg.Append("help"))
						m.Push("inputs", msg.Append("list"))
						m.Push("feature", msg.Append("meta"))
					}
				})
				return
			}

			// 查询命令
			cmds := []string{}
			m.Grows(ice.CHAT_RIVER, prefix, kit.MDB_ID, kit.Format(kit.Int(arg[2])+1), func(index int, value map[string]interface{}) {
				if meta, ok := kit.Value(value, "meta").(map[string]interface{}); ok {
					cmds = kit.Simple(m.Space(meta["pod"]), kit.Keys(meta["ctx"], meta["cmd"]), arg[3:])
				}
			})

			// 执行命令
			if m.Right(cmds) {
				m.Cmdy(cmds).Option("cmds", cmds)
			}
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
