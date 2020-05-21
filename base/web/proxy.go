package web

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_PROXY: {Name: "proxy", Help: "代理", Value: kit.Data(kit.MDB_SHORT, "proxy")},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_PROXY: {Name: "proxy name cmd auto", Help: "代理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "add":
					m.Rich(ice.WEB_SPACE, nil, kit.Dict(
						kit.MDB_TYPE, ice.WEB_BETTER, kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2],
					))
					m.Conf(ice.WEB_PROXY, kit.Keys("meta.better", arg[1]), arg[2])
					m.Logs(ice.LOG_INSERT, "proxy", arg[1], "cb", arg[2])
					return
				}

				m.Richs(ice.WEB_SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
					if value[kit.MDB_TYPE] == ice.WEB_BETTER {
						arg[0] = m.Cmdx(m.Conf(ice.WEB_PROXY, kit.Keys("meta.better", arg[0])))
						m.Logs(ice.LOG_SELECT, "proxy", value["name"], "space", arg[0])
					}
				})

				m.Cmdy(ice.WEB_ROUTE, arg[0], arg[1:])
			}},
		}}, nil)
}
