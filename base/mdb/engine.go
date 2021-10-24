package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const ENGINE = "engine"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ENGINE: {Name: ENGINE, Help: "引擎", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_TYPE, kit.MDB_FIELD, "time,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		ENGINE: {Name: "engine type name text auto", Help: "引擎", Action: ice.MergeAction(map[string]*ice.Action{
			CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(INSERT, m.PrefixKey(), "", HASH, m.OptionSimple("type,name,text"))
			}},
		}, HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				m.Optionv(kit.Keycb(SELECT), func(fields []string, value map[string]interface{}) {
					m.Cmdy(kit.Keys(value[kit.MDB_TEXT], value[kit.MDB_NAME]),
						m.CommandKey(), arg[0], arg[1], kit.Select("", arg, 2), kit.Slice(arg, 3))
				})
			}
			if HashSelect(m, arg...); len(arg) == 0 {
				m.Sort(kit.MDB_TYPE)
			}
		}},
	}})
}
