package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

func _route_travel(m *ice.Message, route string) {
	if route == "" {
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
	} else {
		m.Cmd(SPACE, route, ROUTE).Table(func(index int, value map[string]string, head []string) {
			m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
			m.Push(kit.MDB_ROUTE, kit.Keys(route, value[kit.MDB_ROUTE]))
		})
	}
}

const ROUTE = "route"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ROUTE: {Name: ROUTE, Help: "路由器", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ROUTE)},
		},
		Commands: map[string]*ice.Command{
			ROUTE: {Name: "route route ctx cmd auto 启动 添加", Help: "路由", Action: map[string]*ice.Action{
				"inputs": {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "cmd":
						m.Cmdy(SPACE, m.Option("route"), "command")
					case "name":
						m.Cmdy(SPACE, m.Option("route"), "dream")
					case "template":
						m.Option(nfs.DIR_DEEP, true)
						m.Cmdy(nfs.DIR, "usr/icebergs")
						m.Sort("path")
					}
				}},
				"invite": {Name: "invite", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.publish", "contexts", "tmux")
					m.Cmdy("web.code.publish", "contexts", "base")
					m.Cmdy("web.code.publish", "contexts", "miss")
				}},
				"start": {Name: "start type=worker,server name=hi@key repos", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option("route"), "dream", "start", arg)
					m.Sleep("3s")
				}},
				"gen": {Name: "gen module=hi@key template=usr/icebergs/misc/zsh/zsh.go@key", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option("route"), "web.code.autogen",
						"create", "name", m.Option("module"), "from", m.Option("template"))
				}},
				"stop": {Name: "stop", Help: "结束", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option("route"), "exit")
					m.Sleep("3s")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 3 && arg[3] == "run" {
					m.Cmdy(SPACE, arg[0], kit.Keys(arg[1], arg[2]), arg[4:])
					return
				}
				if len(arg) > 2 && arg[0] != "" {
					// m.Cmdy(SPACE, arg[0], kit.Split(kit.Keys(arg[1], strings.Join(arg[2:], " "))))
					m.Cmdy(SPACE, arg[0], "context", arg[1], "command", arg[2])
					m.Option("_process", "_field")
					m.Option("_prefix", arg[0], arg[1], arg[2], "run")
					return
				}
				if len(arg) > 1 && arg[0] != "" {
					m.Cmd(SPACE, arg[0], "context", arg[1], "command").Table(func(index int, value map[string]string, head []string) {
						m.Push("cmd", value["key"])
						m.Push("name", value["name"])
						m.Push("help", value["help"])
					})
					return
				}
				if len(arg) > 0 && arg[0] != "" {
					m.Cmd(SPACE, arg[0], "context").Table(func(index int, value map[string]string, head []string) {
						m.Push("ctx", kit.Keys(value["ups"], value["name"]))
						m.Push("status", value["status"])
						m.Push("stream", value["stream"])
						m.Push("help", value["help"])
					})
					return
				}

				if _route_travel(m, kit.Select("", arg, 0)); m.W == nil {
					return
				}

				m.Table(func(index int, value map[string]string, field []string) {
					m.PushRender(kit.MDB_LINK, "a", value[kit.MDB_ROUTE],
						kit.MergeURL(m.Option(ice.MSG_USERWEB), "pod", kit.Keys(m.Option("pod", value[kit.MDB_ROUTE]))))
					switch value[kit.MDB_TYPE] {
					case SERVER:
						m.PushRender("action", "button", "创建", "启动")
					case WORKER:
						m.PushRender("action", "button", "创建", "结束")
					}
				})
				m.Sort(kit.MDB_ROUTE)
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
