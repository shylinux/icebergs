package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const RENDER = "render"

func init() {
	Index.MergeCommands(ice.Commands{RENDER: {Name: "render type name text auto", Help: "渲染", Actions: RenderAction()}})
}

func RenderAction(args ...ice.Any) ice.Actions {
	return ice.MergeAction(ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
				cs[m.CommandKey()] = &ice.Config{Value: kit.Data(SHORT, TYPE)}
			}
		}},
		CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
			m.Option(TYPE, kit.Ext(m.Option(TYPE)))
			m.Option(NAME, kit.Select(m.Option(TYPE), m.Option(NAME)))
			m.Cmdy(INSERT, m.PrefixKey(), "", HASH, m.OptionSimple("type,name,text"))
		}},
		SELECT: {Name: "select type name text auto", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 1 {
				for _, k := range kit.Split(arg[0]) {
					HashSelect(m.Spawn(ice.OptionFields("")), k).Tables(func(value ice.Maps) {
						m.Cmdy(kit.Keys(value[TEXT], value[NAME]), m.CommandKey(), k, arg[1], kit.Select("", arg, 2), kit.Slice(arg, 3))
					})
				}
				return
			}
			HashSelect(m, arg...)
		}},
	})
}
