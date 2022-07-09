package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const RENDER = "render"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		RENDER: {Name: RENDER, Help: "渲染", Value: kit.Data(SHORT, TYPE, FIELD, "time,type,name,text")},
	}, Commands: ice.Commands{
		RENDER: {Name: "render type name text auto", Help: "渲染", Actions: ice.Actions{
			CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Option(NAME, kit.Select(m.Option(TYPE), m.Option(NAME)))
				m.Option(TYPE, kit.Ext(m.Option(TYPE)))
				m.Cmdy(INSERT, m.PrefixKey(), "", HASH, m.OptionSimple("type,name,text"))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 1 {
				m.Cmdy(SELECT, m.PrefixKey(), "", HASH, m.Config(SHORT), arg, func(value ice.Map) {
					m.Cmdy(kit.Keys(value[TEXT], value[NAME]), m.CommandKey(), arg[0], arg[1], kit.Select("", arg, 2), kit.Slice(arg, 3))
				})
				return
			}
			if HashSelect(m, arg...); len(arg) == 0 {
				m.Sort(TYPE)
			} else if len(arg) == 1 {
				m.DisplayStoryJSON()
				m.Echo(kit.Formats(m.Confv(m.Append(NAME), "meta.plug")))
			}
		}},
	}})
}
