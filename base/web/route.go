package web

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"strings"
)

func _route_split(arg ...string) (string, string) {
	target, rest := "*", ""
	if len(arg) > 0 {
		ls := strings.SplitN(arg[0], ".", 2)
		if target = ls[0]; len(ls) > 1 {
			rest = ls[1]
		}
	}
	return target, rest
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_ROUTE: {Name: "route", Help: "路由", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_ROUTE: {Name: "route name cmd auto", Help: "路由", Meta: kit.Dict("detail", []string{"分组"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 && arg[0] == "action" {
					switch arg[1] {
					case "group", "分组":
						if m.Option("grp") != "" && m.Option("name") != "" {
							m.Cmdy(ice.WEB_GROUP, m.Option("grp"), "add", m.Option("name"))
						}
					}
					return
				}
				if len(arg) > 0 && arg[0] == "" {
					m.Cmdy(arg[1:])
					return
				}

				target, rest := _route_split(arg...)
				m.Richs(ice.WEB_SPACE, nil, target, func(key string, val map[string]interface{}) {
					if len(arg) > 1 {
						m.Call(false, func(res *ice.Message) *ice.Message { return res })
						ls := []interface{}{ice.WEB_SPACE, val[kit.MDB_NAME]}
						// 发送命令
						if rest != "" {
							ls = append(ls, ice.WEB_SPACE, rest)
						}
						m.Cmdy(ls, arg[1:])
						return
					}

					switch val[kit.MDB_TYPE] {
					case ice.WEB_SERVER:
						if val[kit.MDB_NAME] == m.Conf(ice.CLI_RUNTIME, "node.name") {
							// 避免循环
							return
						}

						// 远程查询
						m.Cmd(ice.WEB_SPACE, val[kit.MDB_NAME], ice.WEB_ROUTE).Table(func(index int, value map[string]string, head []string) {
							m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
							m.Push(kit.MDB_NAME, kit.Keys(val[kit.MDB_NAME], value[kit.MDB_NAME]))
						})
						fallthrough
					case ice.WEB_WORKER:
						// 本机查询
						m.Push(kit.MDB_TYPE, val[kit.MDB_TYPE])
						m.Push(kit.MDB_NAME, val[kit.MDB_NAME])
					}
				})
				if m.W != nil && len(arg) < 2 {
					m.Table(func(index int, value map[string]string, field []string) {
						m.Push("link", Format("a", kit.MergeURL(m.Option(ice.MSG_USERWEB), "pod", value["name"]), value["name"]))
					})
				}
			}},
			"/route/": {Name: "/route/", Help: "路由器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "login":
					if m.Option(ice.MSG_USERNAME) != "" {
						m.Push(ice.MSG_USERNAME, m.Option(ice.MSG_USERNAME))
						m.Info("username: %v", m.Option(ice.MSG_USERNAME))
						break
					}
					if m.Option(ice.MSG_SESSID) != "" && m.Cmdx(ice.AAA_SESS, "check", m.Option(ice.MSG_SESSID)) != "" {
						m.Info("sessid: %v", m.Option(ice.MSG_SESSID))
						break
					}

					sessid := m.Cmdx(ice.AAA_SESS, "create", "")
					share := m.Cmdx(ice.WEB_SHARE, "login", m.Option(ice.MSG_USERIP), sessid)
					Render(m, "cookie", sessid)
					m.Render(share)
				}
			}},
		}}, nil)
}
