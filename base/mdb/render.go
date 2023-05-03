package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const RENDER = "render"

func init() { Index.MergeCommands(ice.Commands{RENDER: {Help: "渲染", Actions: RenderAction()}}) }

func RenderAction(arg ...ice.Any) ice.Actions {
	return ice.MergeActions(ice.Actions{ice.CTX_INIT: AutoConfig(SHORT, TYPE, FIELD, "time,type,name,text", arg),
		CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
			Config(m, SHORT, TYPE)
			HashCreate(m)
		}},
		SELECT: {Name: "select type name text auto create", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 || arg[0] == "" {
				HashSelect(m, arg...)
				return
			}
			m.OptionDefault(ice.MSG_FIELDS, kit.Select("", arg, 2))
			kit.For(kit.Split(arg[0]), func(k string) {
				HashSelects(m.Spawn(), k).Table(func(value ice.Maps) {
					m.Cmdy(kit.Keys(value[TEXT], value[NAME]), m.CommandKey(), k, arg[1], kit.Select("", arg, 2), kit.Slice(arg, 3))
				})
			})
			m.Sort(m.OptionFields())
		}},
	}, ClearOnExitHashAction())
}
