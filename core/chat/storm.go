package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
)

func _storm_list(m *ice.Message, river string) {
	m.Set(ice.MSG_OPTION, kit.MDB_KEY)
	m.Set(ice.MSG_OPTION, kit.MDB_NAME)
	m.Richs(RIVER, kit.Keys(kit.MDB_HASH, river, TOOL), kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		m.Push(key, value[kit.MDB_META], []string{kit.MDB_KEY, kit.MDB_NAME})
	})
	m.Sort(kit.MDB_NAME)
}
func _storm_tool(m *ice.Message, river, storm string, arg ...string) {
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL, kit.MDB_HASH, storm)
	for i := 0; i < len(arg)-3; i += 4 {
		id := m.Grow(RIVER, kit.Keys(prefix), kit.Data(
			POD, arg[i], CTX, arg[i+1], CMD, arg[i+2], "help", arg[i+3],
		))
		m.Log_INSERT(RIVER, river, STORM, storm, kit.MDB_HASH, id, TOOL, arg[i:i+4])
	}
}
func _storm_share(m *ice.Message, river, storm, name string, arg ...string) {
	m.Cmdy(web.SHARE, STORM, name, storm, RIVER, river, arg)
}
func _storm_rename(m *ice.Message, river, storm string, name string) {
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL, kit.MDB_HASH, storm)
	old := m.Conf(RIVER, kit.Keys(prefix, kit.MDB_META, kit.MDB_NAME))
	m.Log_MODIFY(RIVER, river, STORM, storm, kit.MDB_VALUE, name, "old", old)
	m.Conf(RIVER, kit.Keys(prefix, kit.MDB_META, kit.MDB_NAME), name)
}
func _storm_remove(m *ice.Message, river string, storm string) {
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL)
	m.Richs(RIVER, kit.Keys(prefix), storm, func(value map[string]interface{}) {
		m.Log_REMOVE(RIVER, river, STORM, storm, kit.MDB_VALUE, kit.Format(value))
	})
	m.Conf(RIVER, kit.Keys(prefix, kit.MDB_HASH, storm), "")
}

const (
	POD = "pod"
	CTX = "ctx"
	CMD = "cmd"
	ARG = "arg"
	VAL = "val"
)

const STORM = "storm"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/" + STORM: {Name: "/storm", Help: "暴风雨", Action: map[string]*ice.Action{
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					_storm_remove(m, m.Option(RIVER), m.Option(STORM))
				}},
				mdb.RENAME: {Name: "rename name", Help: "重命名", Hand: func(m *ice.Message, arg ...string) {
					_storm_rename(m, m.Option(RIVER), m.Option(STORM), arg[0])
				}},
				web.SHARE: {Name: "share name", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					_storm_share(m, m.Option(RIVER), m.Option(STORM), arg[0])
				}},
				TOOL: {Name: "tool [pod ctx cmd help]...", Help: "添加工具", Hand: func(m *ice.Message, arg ...string) {
					_storm_tool(m, m.Option(RIVER), m.Option(STORM), arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_storm_list(m, m.Option(RIVER))
			}},
		},
	}, nil)
}
