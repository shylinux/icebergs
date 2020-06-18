package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

func _storm_list(m *ice.Message, river string) {
	m.Richs(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, river, "tool"), "*", func(key string, value map[string]interface{}) {
		m.Push(key, value["meta"], []string{kit.MDB_KEY, kit.MDB_NAME})
	})
	m.Sort(kit.MDB_NAME)
}
func _storm_tool(m *ice.Message, river, storm string, arg ...string) {
	prefix := kit.Keys(kit.MDB_HASH, river, "tool", kit.MDB_HASH, storm)
	for i := 0; i < len(arg)-3; i += 4 {
		id := m.Grow(ice.CHAT_RIVER, kit.Keys(prefix), kit.Data(
			"pod", arg[i], "ctx", arg[i+1], "cmd", arg[i+2], "help", arg[i+3],
		))
		m.Log_INSERT(RIVER, river, STORM, storm, "hash", id, "tool", arg[i:i+4])
	}
}
func _storm_share(m *ice.Message, river, storm, name string, arg ...string) {
	m.Cmdy(ice.WEB_SHARE, ice.TYPE_STORM, name, storm, RIVER, river, arg)
}
func _storm_rename(m *ice.Message, river, storm string, name string) {
	prefix := kit.Keys(kit.MDB_HASH, river, "tool", kit.MDB_HASH, storm)
	old := m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_META, kit.MDB_NAME))
	m.Log_MODIFY(RIVER, river, STORM, storm, "value", name, "old", old)
	m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_META, kit.MDB_NAME), name)
}
func _storm_remove(m *ice.Message, river string, storm string) {
	prefix := kit.Keys(kit.MDB_HASH, river, "tool")
	m.Richs(ice.CHAT_RIVER, kit.Keys(prefix), storm, func(value map[string]interface{}) {
		m.Log_REMOVE(RIVER, river, STORM, storm, "value", kit.Format(value))
	})
	m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, storm), "")
}

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/storm": {Name: "/storm share|tool|rename|remove arg...", Help: "暴风雨", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "share":
				// TODO check right
				_storm_share(m, m.Option(RIVER), m.Option(STORM), m.Option("name"))
				return
			}

			if m.Warn(m.Option(ice.MSG_RIVER) == "", "not join") {
				// m.Render("status", 402, "not join")
				return
			}

			if len(arg) < 3 {
				_storm_list(m, arg[0])
				return
			}

			if !m.Right(cmd, arg[2]) {
				m.Render("status", 403, "not auth")
				return
			}

			switch arg[2] {
			case "add", "tool":
				_storm_tool(m, arg[0], arg[1], arg[3:]...)
			case "rename":
				_storm_rename(m, arg[0], arg[1], arg[3])
			case "remove":
				_storm_remove(m, arg[0], arg[1])
			}
		}},
	}}, nil)
}
