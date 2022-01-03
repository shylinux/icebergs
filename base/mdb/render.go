package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const RENDER = "render"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		RENDER: {Name: "render", Help: "渲染", Value: kit.Data(SHORT, TYPE, FIELD, "time,type,name,text")},
	}, Commands: map[string]*ice.Command{
		RENDER: {Name: "render type name text auto", Help: "渲染", Action: map[string]*ice.Action{
			CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(INSERT, m.PrefixKey(), "", HASH, m.OptionSimple("type,name,text"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				m.Optionv(kit.Keycb(SELECT), func(fields []string, value map[string]interface{}) {
					m.Cmdy(kit.Keys(value[TEXT], value[NAME]),
						m.CommandKey(), arg[0], arg[1], kit.Select("", arg, 2), kit.Slice(arg, 3))
				})
			}
			if HashSelect(m, arg...); len(arg) == 0 {
				m.Sort(TYPE)
			}
		}},
	}})
}
