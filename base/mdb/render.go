package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const RENDER = "render"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			RENDER: {Name: "render", Help: "渲染", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TYPE)},
		},
		Commands: map[string]*ice.Command{
			RENDER: {Name: "render type name text auto", Help: "渲染", Action: map[string]*ice.Action{
				CREATE: {Name: "create type cmd ctx", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Rich(RENDER, nil, kit.Dict(kit.MDB_TYPE, arg[0], kit.MDB_NAME, kit.Select(arg[0], arg, 1), kit.MDB_TEXT, kit.Select("", arg, 2)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || arg[0] == "" {
					m.Richs(RENDER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						m.Push(key, value, []string{kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
					})
					return
				}

				m.Richs(RENDER, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Cmdy(kit.Keys(value[kit.MDB_TEXT], value[kit.MDB_NAME]), RENDER, arg[0], arg[1], kit.Select("", arg, 2))
				})
			}},
		}})
}
