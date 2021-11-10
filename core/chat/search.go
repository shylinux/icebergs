package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		"search": {Name: "search", Help: "搜索", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
	}, Commands: map[string]*ice.Command{
		"/search": {Name: "/search", Help: "搜索引擎", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				m.Richs("/search", "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					if value = kit.GetMeta(value); arg[1] != "" && !kit.Contains(value[kit.MDB_NAME], arg[1]) {
						return
					}
					m.PushSearch(ice.CMD, "/search", value)
				})
			}},
			mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(m.Space(m.Option(ice.POD)), mdb.RENDER, arg[1:])
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if kit.Contains(arg[1], ";") {
				arg = kit.Split(arg[1], ";", ";", ";")
			}
			defer m.StatusTimeCount()
			if m.Cmdy(m.Space(m.Option(ice.POD)), mdb.SEARCH, arg); arg[1] == "" {
				return
			}
			m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH,
				kit.MDB_NAME, arg[1], kit.MDB_TYPE, arg[0], kit.MDB_TEXT, kit.Select("", arg, 2))
		}},
	}})
}
