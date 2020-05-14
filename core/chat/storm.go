package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/storm": {Name: "/storm", Help: "暴风雨", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "share":
				m.Cmdy(ice.WEB_SHARE, ice.TYPE_STORM, m.Option("name"), m.Option("storm"), "river", m.Option("river"))
				return
			}

			if m.Warn(m.Option(ice.MSG_RIVER) == "", "not join") {
				// m.Render("status", 402, "not join")
				return
			}

			prefix := kit.Keys(kit.MDB_HASH, arg[0], "tool")
			if len(arg) < 3 {
				// 应用列表
				m.Richs(ice.CHAT_RIVER, prefix, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"], []string{kit.MDB_KEY, kit.MDB_NAME})
				})
				m.Log(ice.LOG_SELECT, "%s", m.Format("append"))
				m.Sort(kit.MDB_NAME)
				return
			}

			if !m.Right(cmd, arg[2]) {
				m.Render("status", 403, "not auth")
				return
			}

			switch arg[2] {
			case "add":
				// 添加命令
				for i := 3; i < len(arg)-3; i += 4 {
					id := m.Grow(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1]), kit.Data(
						"pod", arg[i], "ctx", arg[i+1], "cmd", arg[i+2], "help", arg[i+3],
					))
					m.Log(ice.LOG_INSERT, "storm: %s %d: %v", arg[1], id, arg[i:i+4])
				}

			case "rename":
				// 重命名应用
				old := m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1], kit.MDB_META, kit.MDB_NAME))
				m.Log(ice.LOG_MODIFY, "storm: %s %s->%s", arg[1], old, arg[3])
				m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1], kit.MDB_META, kit.MDB_NAME), arg[3])

			case "remove":
				// 删除应用
				m.Richs(ice.CHAT_RIVER, kit.Keys(prefix), arg[1], func(value map[string]interface{}) {
					m.Log(ice.LOG_REMOVE, "storm: %s value: %s", arg[1], kit.Format(value))
				})
				m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1]), "")
			}
		}},
	}}, nil)
}
