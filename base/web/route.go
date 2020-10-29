package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
)

func _route_travel(m *ice.Message, route string) {
	if route == "" {
		m.Richs(SPACE, nil, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
			switch val[kit.MDB_TYPE] {
			case SERVER:
				if val[kit.MDB_NAME] == ice.Info.NodeName {
					return // 避免循环
				}

				// 远程查询
				m.Cmd(SPACE, val[kit.MDB_NAME], ROUTE).Table(func(index int, value map[string]string, head []string) {
					m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
					m.Push(kit.SSH_ROUTE, kit.Keys(val[kit.MDB_NAME], value[kit.SSH_ROUTE]))
				})
				fallthrough
			case WORKER:
				// 本机查询
				m.Push(kit.MDB_TYPE, val[kit.MDB_TYPE])
				m.Push(kit.SSH_ROUTE, val[kit.MDB_NAME])
			}
		})
	} else {
		m.Cmd(SPACE, route, ROUTE).Table(func(index int, value map[string]string, head []string) {
			m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
			m.Push(kit.SSH_ROUTE, kit.Keys(route, value[kit.SSH_ROUTE]))
		})
	}
}

const ROUTE = "route"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ROUTE: {Name: ROUTE, Help: "路由器", Value: kit.Data(kit.MDB_SHORT, kit.SSH_ROUTE)},
		},
		Commands: map[string]*ice.Command{
			ROUTE: {Name: "route route ctx cmd auto invite share", Help: "路由", Action: map[string]*ice.Action{
				SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					h := m.Cmdx(SHARE, mdb.CREATE, kit.MDB_TYPE, "login")
					p := kit.MergeURL(m.Option(ice.MSG_USERWEB), SHARE, h)
					m.Cmdy("web.wiki.spark", "shell", p).Render("")
					m.Cmdy("web.wiki.image", "qrcode", p)
				}},
				mdb.INVITE: {Name: "invite", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range []string{"tmux", "base", "miss"} {
						m.Cmdy("web.code.publish", "contexts", k)
						m.Cmdy("web.wiki.spark")
					}

					m.Cmdy("web.wiki.spark", "shell", m.Option(ice.MSG_USERWEB))
					m.Cmdy("web.wiki.image", "qrcode", m.Option(ice.MSG_USERWEB))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "cmd":
						m.Cmdy(SPACE, m.Option(ROUTE), ctx.COMMAND)
					case "name":
						m.Cmdy(SPACE, m.Option(ROUTE), DREAM)
					case "template":
						m.Option(nfs.DIR_DEEP, true)
						m.Cmdy(nfs.DIR, "usr/icebergs")
						m.Sort(kit.MDB_PATH)
					}
				}},
				gdb.START: {Name: "start type=worker,server repos name", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option(ROUTE), DREAM, gdb.START, arg)
				}},
				gdb.STOP: {Name: "stop", Help: "结束", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option(ROUTE), "exit")
					m.Sleep("3s")
				}},
				"gen": {Name: "gen module=hi@key template=usr/icebergs/misc/zsh/zsh.go@key", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option(ROUTE), "web.code.autogen",
						"create", "name", m.Option("module"), "from", m.Option("template"))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || arg[0] == "" {
					if _route_travel(m, kit.Select("", arg, 0)); m.W == nil {
						return
					}

					m.Table(func(index int, value map[string]string, field []string) {
						if value[kit.MDB_TYPE] != MYSELF {
							m.PushRender(kit.MDB_LINK, "a", value[kit.SSH_ROUTE],
								kit.MergeURL(m.Option(ice.MSG_USERWEB), kit.SSH_POD, kit.Keys(m.Option(kit.SSH_POD, value[kit.SSH_ROUTE]))))
						}

						switch value[kit.MDB_TYPE] {
						case SERVER:
							m.PushButton(gdb.START)
						case WORKER:
							m.PushButton(gdb.STOP)
						default:
							m.PushButton("")
						}
					})

					m.Cmd(tcp.HOST).Table(func(index int, value map[string]string, head []string) {
						m.Push(kit.MDB_TYPE, MYSELF)
						m.Push(kit.SSH_ROUTE, ice.Info.NodeName)
						u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
						m.PushRender(kit.MDB_LINK, "a", value["ip"], kit.Format("%s://%s:%s", u.Scheme, value["ip"], u.Port()))
						m.PushButton(gdb.START)
					})

					m.Push(kit.MDB_TYPE, MYSELF)
					m.Push(kit.SSH_ROUTE, ice.Info.NodeName)
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					m.PushRender(kit.MDB_LINK, "a", "localhost", kit.Format("%s://%s:%s", u.Scheme, "localhost", u.Port()))
					m.PushButton(gdb.START)

					m.Sort(kit.SSH_ROUTE)
					return // 设备列表
				}

				if len(arg) > 3 { // 执行命令
					m.Cmdy(SPACE, arg[0], kit.Keys(arg[1], arg[2]), arg[4:])

				} else if len(arg) > 2 { // 加载插件
					m.Cmdy(SPACE, arg[0], ctx.CONTEXT, arg[1], ctx.COMMAND, arg[2])
					m.Option("_prefix", arg[0], arg[1], arg[2], "run")
					m.Option("_process", "_field")

				} else if len(arg) > 1 { // 命令列表
					m.Cmd(SPACE, arg[0], ctx.CONTEXT, arg[1], ctx.COMMAND).Table(func(index int, value map[string]string, head []string) {
						m.Push("cmd", value["key"])
						m.Push("name", value["name"])
						m.Push("help", value["help"])
					})

				} else if len(arg) > 0 { // 模块列表
					m.Cmd(SPACE, arg[0], ctx.CONTEXT).Table(func(index int, value map[string]string, head []string) {
						m.Push("ctx", kit.Keys(value["ups"], value["name"]))
						m.Push("status", value["status"])
						m.Push("stream", value["stream"])
						m.Push("help", value["help"])
					})
				}
			}},
		}})
}
