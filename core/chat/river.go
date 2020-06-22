package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
)

func _river_list(m *ice.Message) {
	m.Set(ice.MSG_OPTION, kit.MDB_KEY)
	m.Set(ice.MSG_OPTION, kit.MDB_NAME)
	m.Richs(RIVER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, USER), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
			m.Push(key, value[kit.MDB_META], []string{kit.MDB_KEY, kit.MDB_NAME})
		})
	})
}
func _river_user(m *ice.Message, river string, user ...string) {
	prefix := kit.Keys(kit.MDB_HASH, river, USER)
	m.Rich(RIVER, prefix, kit.Data(aaa.USERNAME, cli.UserName))
	for _, v := range user {
		m.Rich(RIVER, prefix, kit.Data(aaa.USERNAME, v))
		m.Log_INSERT(RIVER, river, USER, v)
	}
}
func _river_share(m *ice.Message, river, name string, arg ...string) {
	m.Cmdy(web.SHARE, RIVER, name, river, arg)
}
func _river_rename(m *ice.Message, river string, name string) {
	prefix := kit.Keys(kit.MDB_HASH, river, kit.MDB_META, kit.MDB_NAME)
	old := m.Conf(RIVER, prefix)
	m.Log_MODIFY(RIVER, river, kit.MDB_VALUE, name, "old", old)
	m.Conf(RIVER, prefix, name)
}
func _river_remove(m *ice.Message, river string) {
	m.Richs(RIVER, nil, river, func(value map[string]interface{}) {
		m.Log_REMOVE(RIVER, river, kit.MDB_VALUE, kit.Format(value))
	})
	m.Conf(RIVER, kit.Keys(kit.MDB_HASH, river), "")
}

const (
	USER = "user"
	TOOL = "tool"
)
const RIVER = "river"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/" + RIVER: {Name: "/river", Help: "小河流",
			Action: map[string]*ice.Action{
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					_river_remove(m, m.Option(RIVER))
				}},
				mdb.RENAME: {Name: "rename name", Help: "重命名", Hand: func(m *ice.Message, arg ...string) {
					_river_rename(m, m.Option(RIVER), arg[0])
				}},
				web.SHARE: {Name: "share name", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					_river_share(m, m.Option(RIVER), arg[0])
				}},
				USER: {Name: "user user...", Help: "添加用户", Hand: func(m *ice.Message, arg ...string) {
					_river_user(m, m.Option(RIVER), arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_river_list(m)
			}},
	}}, nil)
}
