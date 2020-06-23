package mdb

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const ENGINE = "engine"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ENGINE: {Name: "engine", Help: "引擎", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TYPE)},
		},
		Commands: map[string]*ice.Command{
			ENGINE: {Name: "engine type name text arg...", Help: "引擎", Action: map[string]*ice.Action{
				CREATE: {Name: "create type name [text]", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Rich(ENGINE, nil, kit.Dict(kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, kit.Select("", arg, 2)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Richs(ENGINE, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Cmdy(kit.Keys(value[kit.MDB_TEXT], value[kit.MDB_NAME]), arg)
				}) == nil {
					m.Cmdy(arg[0], ENGINE, arg)
				}
			}},
		}}, nil)
}
