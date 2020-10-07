package web

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const PROXY = "proxy"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PROXY: {Name: "proxy", Help: "代理", Value: kit.Data(kit.MDB_SHORT, PROXY)},
		},
		Commands: map[string]*ice.Command{
			PROXY: {Name: "proxy name cmd auto", Help: "代理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "add":
					m.Rich(SPACE, nil, kit.Dict(
						kit.MDB_TYPE, BETTER, kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2],
					))
					m.Conf(PROXY, kit.Keys("meta.better", arg[1]), arg[2])
					m.Logs(ice.LOG_INSERT, "proxy", arg[1], "cb", arg[2])
					return
				}

				m.Richs(SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
					if value[kit.MDB_TYPE] == BETTER {
						arg[0] = m.Cmdx(m.Conf(PROXY, kit.Keys("meta.better", arg[0])))
						m.Logs(ice.LOG_SELECT, "proxy", value["name"], "space", arg[0])
					}
				})

				m.Cmdy(ROUTE, arg[0], arg[1:])
			}},
		}}, nil)
}
