package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const PLUGIN = "plugin"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		PLUGIN: {Name: PLUGIN, Help: "插件", Value: kit.Data(SHORT, TYPE, FIELD, "time,type,name,text")},
	}, Commands: map[string]*ice.Command{
		PLUGIN: {Name: "plugin type name text auto", Help: "插件", Action: map[string]*ice.Action{
			CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Option(NAME, kit.Select(m.Option(TYPE), m.Option(NAME)))
				m.Option(TYPE, kit.Ext(m.Option(TYPE)))
				m.Cmdy(INSERT, m.PrefixKey(), "", HASH, m.OptionSimple("type,name,text"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				m.Cmdy(SELECT, m.PrefixKey(), "", HASH, m.Config(SHORT), arg, func(value map[string]interface{}) {
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
