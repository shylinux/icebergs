package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const ENGINE = "engine"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ENGINE: {Name: "engine", Help: "引擎", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TYPE)},
	}, Commands: map[string]*ice.Command{
		ENGINE: {Name: "engine type name text auto", Help: "引擎", Action: map[string]*ice.Action{
			CREATE: {Name: "create type cmd ctx", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Log_CREATE(ENGINE, arg[0], kit.MDB_NAME, kit.Select(arg[0], arg, 1))
				m.Rich(ENGINE, nil, kit.Dict(kit.MDB_TYPE, arg[0], kit.MDB_NAME, kit.Select(arg[0], arg, 1), kit.MDB_TEXT, kit.Select("", arg, 2)))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				m.Richs(ENGINE, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
				})
				return
			}

			if len(arg) == 2 {
				arg = append(arg, "")
			}
			m.Richs(ENGINE, nil, arg[0], func(key string, value map[string]interface{}) {
				m.Cmdy(kit.Keys(value[kit.MDB_TEXT], value[kit.MDB_NAME]), ENGINE, arg[0], arg[1], arg[2], arg[3:])
			})
		}},
	}})
}
