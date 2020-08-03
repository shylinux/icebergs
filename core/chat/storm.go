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
	m.Set(ice.MSG_OPTION, kit.MDB_COUNT)

	if p := m.Option(POD); p != "" {
		m.Option(POD, "")
		// 代理列表
		m.Cmdy(web.SPACE, p, "web.chat./storm")
	}
	ok := true
	m.Richs(RIVER, kit.Keys(kit.MDB_HASH, river, USER), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
		ok = true
	})
	m.Richs(RIVER, kit.Keys(kit.MDB_HASH, river), m.Option(ice.MSG_RIVER), func(k string, val map[string]interface{}) {
		ok = true
	})
	if ok {
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, river, TOOL), kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			m.Push(key, value[kit.MDB_META], []string{kit.MDB_KEY, kit.MDB_NAME, kit.MDB_COUNT})
		})
		m.Sort(kit.MDB_NAME)
	}
}
func _storm_tool(m *ice.Message, river, storm string, arg ...string) { // pod ctx cmd help
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL, kit.MDB_HASH, storm)
	for i := 0; i < len(arg)-3; i += 4 {
		m.Grow(RIVER, kit.Keys(prefix), kit.Data(
			POD, arg[i], CTX, arg[i+1], CMD, arg[i+2], "help", arg[i+3],
		))
		m.Log_INSERT(RIVER, river, STORM, storm, TOOL, arg[i:i+4])
	}
}
func _storm_share(m *ice.Message, river, storm, name string, arg ...string) {
	m.Cmdy(web.SHARE, STORM, name, storm, RIVER, river, arg)
}
func _storm_remove(m *ice.Message, river string, storm string) {
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL)
	m.Richs(RIVER, kit.Keys(prefix), storm, func(value map[string]interface{}) {
		m.Log_REMOVE(RIVER, river, STORM, storm, kit.MDB_VALUE, kit.Format(value))
	})
	m.Conf(RIVER, kit.Keys(prefix, kit.MDB_HASH, storm), "")
}
func _storm_rename(m *ice.Message, river, storm string, name string) {
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL, kit.MDB_HASH, storm)
	old := m.Conf(RIVER, kit.Keys(prefix, kit.MDB_META, kit.MDB_NAME))
	m.Log_MODIFY(RIVER, river, STORM, storm, kit.MDB_VALUE, name, "old", old)
	m.Conf(RIVER, kit.Keys(prefix, kit.MDB_META, kit.MDB_NAME), name)
}
func _storm_create(m *ice.Message, river string, kind, name, text string, arg ...string) string {
	h := m.Rich(RIVER, kit.Keys(kit.MDB_HASH, river, TOOL), kit.Dict(
		kit.MDB_META, kit.Dict(
			kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
			kit.MDB_EXTRA, kit.Dict(arg),
		),
	))
	m.Log_CREATE(kit.MDB_META, STORM, RIVER, river, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
	m.Echo(h)
	return h
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
		Configs: map[string]*ice.Config{
			STORM: {Name: "storm", Help: "应用", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			"/" + STORM: {Name: "/storm", Help: "暴风雨", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type name text arg...", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_storm_create(m, m.Option(ice.MSG_RIVER), arg[0], arg[1], arg[2], arg[3:]...)
				}},
				mdb.RENAME: {Name: "rename name", Help: "更名", Hand: func(m *ice.Message, arg ...string) {
					_storm_rename(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), arg[0])
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					_storm_remove(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM))
				}},
				web.SHARE: {Name: "share name", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					_storm_share(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), arg[0])
				}},
				"save": {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					prefix := kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(ice.MSG_STORM))
					for i, v := range arg {
						args := kit.Parse(nil, "", kit.Split(v)...)
						m.Conf(RIVER, kit.Keys(prefix, kit.MDB_LIST, i, kit.MDB_META, "args"), args)
						m.Debug("fuck %v %v", kit.Keys(prefix, kit.MDB_LIST, i), args)
					}
				}},
				TOOL: {Name: "tool [pod ctx cmd help]...", Help: "添加工具", Hand: func(m *ice.Message, arg ...string) {
					_storm_tool(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_storm_list(m, m.Option(ice.MSG_RIVER))
			}},
		},
	}, nil)
}
