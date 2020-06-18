package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

func _river_right(m *ice.Message, action string) bool {
	if m.Warn(m.Option(ice.MSG_RIVER) == "", "not join") {
		m.Render("status", 402, "not join")
		return false
	}

	if !m.Right(RIVER, action) {
		m.Render("status", 403, "not auth")
		return false
	}
	return true
}

func _river_list(m *ice.Message) {
	m.Set(ice.MSG_OPTION, "key")
	m.Set(ice.MSG_OPTION, "name")
	m.Richs(RIVER, nil, "*", func(key string, value map[string]interface{}) {
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, "user"), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
			m.Push(key, value["meta"], []string{kit.MDB_KEY, kit.MDB_NAME})
		})
	})
}
func _river_user(m *ice.Message, river string, user ...string) {
	m.Rich(RIVER, kit.Keys(kit.MDB_HASH, river, "user"), kit.Data("username", m.Conf(ice.CLI_RUNTIME, "boot.username")))
	for _, v := range user {
		user := m.Rich(RIVER, kit.Keys(kit.MDB_HASH, river, "user"), kit.Data("username", v))
		m.Log_INSERT(RIVER, river, "hash", user, "user", v)
	}
}
func _river_rename(m *ice.Message, river string, name string) {
	old := m.Conf(RIVER, kit.Keys(kit.MDB_HASH, river, kit.MDB_META, kit.MDB_NAME))
	m.Log_MODIFY(RIVER, river, "value", name, "old", old)
	m.Conf(RIVER, kit.Keys(kit.MDB_HASH, river, kit.MDB_META, kit.MDB_NAME), name)
}
func _river_remove(m *ice.Message, river string) {
	m.Richs(RIVER, nil, river, func(value map[string]interface{}) {
		m.Log_REMOVE(RIVER, river, "value", kit.Format(value))
	})
	m.Conf(RIVER, kit.Keys(kit.MDB_HASH, river), "")
}

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/river": {Name: "/river river user|rename|remove arg...", Help: "小河流",
			Action: map[string]*ice.Action{
				"user": {Name: "user user...", Help: "添加用户", Hand: func(m *ice.Message, arg ...string) {
					if _river_right(m, "user") {
						_river_user(m, m.Option(ice.CHAT_RIVER), arg...)
					}
				}},
				"rename": {Name: "rename name", Help: "重命名", Hand: func(m *ice.Message, arg ...string) {
					if _river_right(m, "rename") {
						_river_rename(m, m.Option(ice.CHAT_RIVER), arg[0])
					}
				}},
				"remove": {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					if _river_right(m, "remove") {
						_river_remove(m, m.Option(ice.CHAT_RIVER))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_river_list(m)
			}},
	}}, nil)
}
