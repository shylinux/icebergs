package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
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

const ROUTE = "route"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ROUTE: {Name: "route", Help: "路由", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			ROUTE: {Name: "route name cmd auto", Help: "路由", Meta: kit.Dict("detail", []string{"分组"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 && arg[0] == "action" {
					switch arg[1] {
					case "group", "分组":
						if m.Option("grp") != "" && m.Option("name") != "" {
							m.Cmdy(GROUP, m.Option("grp"), "add", m.Option("name"))
						}
					}
					return
				}
				if len(arg) > 0 && arg[0] == "" {
					m.Cmdy(arg[1:])
					return
				}

				target, rest := _route_split(arg...)
				m.Richs(SPACE, nil, target, func(key string, val map[string]interface{}) {
					if len(arg) > 1 {
						m.Call(false, func(res *ice.Message) *ice.Message { return res })
						ls := []interface{}{SPACE, val[kit.MDB_NAME]}
						// 发送命令
						if rest != "" {
							ls = append(ls, SPACE, rest)
						}
						m.Cmdy(ls, arg[1:])
						return
					}

					switch val[kit.MDB_TYPE] {
					case SERVER:
						if val[kit.MDB_NAME] == m.Conf(cli.RUNTIME, "node.name") {
							// 避免循环
							return
						}

						// 远程查询
						m.Cmd(SPACE, val[kit.MDB_NAME], ROUTE).Table(func(index int, value map[string]string, head []string) {
							m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
							m.Push(kit.MDB_NAME, kit.Keys(val[kit.MDB_NAME], value[kit.MDB_NAME]))
						})
						fallthrough
					case WORKER:
						// 本机查询
						m.Push(kit.MDB_TYPE, val[kit.MDB_TYPE])
						m.Push(kit.MDB_NAME, val[kit.MDB_NAME])
					}
				})
				if m.W != nil && len(arg) < 2 {
					m.Table(func(index int, value map[string]string, field []string) {
						m.Push("link", m.Cmdx(mdb.RENDER, RENDER.A, value["name"], kit.MergeURL(m.Option(ice.MSG_USERWEB), "pod", value["name"])))
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
					if m.Option(ice.MSG_SESSID) != "" && m.Cmdx(aaa.SESS, "check", m.Option(ice.MSG_SESSID)) != "" {
						m.Info("sessid: %v", m.Option(ice.MSG_SESSID))
						break
					}

					sessid := m.Cmdx(aaa.SESS, "create", "")
					share := m.Cmdx(SHARE, "login", m.Option(ice.MSG_USERIP), sessid)
					Render(m, "cookie", sessid)
					m.Render(share)
				}
			}},
		}}, nil)
}
