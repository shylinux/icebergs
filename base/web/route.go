package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const ROUTE = "route"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ROUTE: {Name: ROUTE, Help: "路由", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ROUTE)},
		},
		Commands: map[string]*ice.Command{
			ROUTE: {Name: "route route auto 启动 创建 邀请", Help: "路由", Action: map[string]*ice.Action{
				"create": {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.install", "contexts", kit.Select("base", m.Option("type")))
				}},
				"start": {Name: "start type=server,worker name@key", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if p := m.Option("route"); p != "" {
						m.Option("route", "")
						m.Cmdy(SPACE, p, "route", "create", arg)
						return
					}
					m.Cmdy("dream", "start", m.Option("name"))
					m.Sleep("3s")
				}},
				"stop": {Name: "stop", Help: "结束", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option("route"), "exit")
				}},
				"inputs": {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option("route"), "dream")
				}},
				"invite": {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.wiki.image", "qrcode", kit.MergeURL(m.Option(ice.MSG_USERWEB), "river", m.Option(ice.MSG_RIVER)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Cmd(SPACE, arg[0], ROUTE).Table(func(index int, value map[string]string, head []string) {
						m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
						m.Push(kit.MDB_ROUTE, kit.Keys(arg[0], value[kit.MDB_ROUTE]))
					})
					return
				}

				m.Richs(SPACE, nil, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
					switch val[kit.MDB_TYPE] {
					case SERVER:
						if val[kit.MDB_NAME] == cli.NodeName {
							// 避免循环
							return
						}

						// 远程查询
						m.Cmd(SPACE, val[kit.MDB_NAME], ROUTE).Table(func(index int, value map[string]string, head []string) {
							m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
							m.Push(kit.MDB_ROUTE, kit.Keys(val[kit.MDB_NAME], value[kit.MDB_ROUTE]))
						})
						fallthrough
					case WORKER:
						// 本机查询
						m.Push(kit.MDB_TYPE, val[kit.MDB_TYPE])
						m.Push(kit.MDB_ROUTE, val[kit.MDB_NAME])
					}
				})
				if m.W == nil {
					return
				}

				m.Table(func(index int, value map[string]string, field []string) {
					m.Push(kit.MDB_LINK, m.Cmdx(mdb.RENDER, RENDER.A, value[kit.MDB_ROUTE],
						kit.MergeURL(m.Option(ice.MSG_USERWEB), "pod", value[kit.MDB_ROUTE])))
					switch value[kit.MDB_TYPE] {
					case SERVER:
						m.Push("action", m.Cmdx(mdb.RENDER, RENDER.Button, "启动"))
					case WORKER:
						m.Push("action", m.Cmdx(mdb.RENDER, RENDER.Button, "结束"))
					}
				})
				m.Sort("route")
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
