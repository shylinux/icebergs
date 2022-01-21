package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const SEARCH = "search"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SEARCH: {Name: SEARCH, Help: "搜索", Value: kit.Data(SHORT, TYPE, FIELD, "time,type,name,text")},
	}, Commands: map[string]*ice.Command{
		SEARCH: {Name: "search type name text auto", Help: "搜索", Action: map[string]*ice.Action{
			CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(INSERT, m.PrefixKey(), "", HASH, m.OptionSimple("type,name,text"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				m.Cmdy(SELECT, m.PrefixKey(), "", HASH, m.Config(SHORT), arg, func(value map[string]interface{}) {
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
