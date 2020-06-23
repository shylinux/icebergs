package mdb

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const RENDER = "render"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			RENDER: {Name: "render", Help: "渲染引擎", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TYPE)},
		},
		Commands: map[string]*ice.Command{
			RENDER: {Name: "search type name text arg...", Help: "渲染引擎", Action: map[string]*ice.Action{
				CREATE: {Name: "create type name [text]", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Rich(RENDER, nil, kit.Dict(kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, kit.Select("", arg, 2)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(RENDER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					m.Cmdy(kit.Keys(value[kit.MDB_TEXT], value[kit.MDB_NAME]), arg[0], arg[1], kit.Select("", arg, 2))
				})
			}},
		}}, nil)
}
