package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const SEARCH = "search"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		SEARCH: {Name: SEARCH, Help: "搜索", Value: kit.Data(SHORT, TYPE, FIELD, "time,type,name,text")},
	}, Commands: ice.Commands{
		SEARCH: {Name: "search type name text auto", Help: "搜索", Actions: ice.Actions{
			CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Option(NAME, kit.Select(m.Option(TYPE), m.Option(NAME)))
				m.Option(TYPE, kit.Ext(m.Option(TYPE)))
				m.Cmdy(INSERT, m.PrefixKey(), "", HASH, m.OptionSimple("type,name,text"))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 1 {
				m.Cmdy(SELECT, m.PrefixKey(), "", HASH, m.Config(SHORT), kit.Slice(arg, 0, 1), func(value ice.Map) {
					m.OptionFields(kit.Select("ctx,cmd,type,name,text", kit.Select(m.OptionFields())))
					m.Cmdy(kit.Keys(value[TEXT], value[NAME]), m.CommandKey(), arg[0], arg[1], kit.Select("", arg, 2), kit.Slice(arg, 3))
				})
				return
			}
			if HashSelect(m, arg...); len(arg) == 0 {
				m.Sort(TYPE)
			}
		}},
	}})
}
