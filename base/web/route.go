package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
)

func _route_travel(m *ice.Message, route string) {
	m.Richs(SPACE, nil, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		switch val[kit.MDB_TYPE] {
		case SERVER: // 远程查询
			if val[kit.MDB_NAME] == ice.Info.NodeName {
				return // 避免循环
			}

			m.Cmd(SPACE, val[kit.MDB_NAME], ROUTE).Table(func(index int, value map[string]string, head []string) {
				m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
				m.Push(kit.SSH_ROUTE, kit.Keys(val[kit.MDB_NAME], value[kit.SSH_ROUTE]))
			})

			fallthrough
		case WORKER: // 本机查询
			m.Push(kit.MDB_TYPE, val[kit.MDB_TYPE])
			m.Push(kit.SSH_ROUTE, val[kit.MDB_NAME])
		}
	})
}
func _route_list(m *ice.Message) {
	// 链接操作
	m.Table(func(index int, value map[string]string, field []string) {
		m.PushAnchor(value[kit.SSH_ROUTE], kit.MergeURL(m.Option(ice.MSG_USERWEB),
			kit.SSH_POD, kit.Keys(m.Option(ice.MSG_USERPOD), value[kit.SSH_ROUTE])))

		switch value[kit.MDB_TYPE] {
		case WORKER:
			m.PushButton(mdb.CREATE)
		case SERVER:
			m.PushButton(tcp.START)
		default:
			m.PushButton("")
		}
	})

	// 网卡信息
	u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
	m.Cmd(tcp.HOST).Table(func(index int, value map[string]string, head []string) {
		m.Push(kit.MDB_TYPE, MYSELF)
		m.Push(kit.SSH_ROUTE, ice.Info.NodeName)
		m.PushAnchor(value[tcp.IP], kit.Format("%s://%s:%s", u.Scheme, value[tcp.IP], u.Port()))
		m.PushButton(tcp.START)
	})

	// 本机信息
	m.Push(kit.MDB_TYPE, MYSELF)
	m.Push(kit.SSH_ROUTE, ice.Info.NodeName)
	m.PushAnchor(tcp.LOCALHOST, kit.Format("%s://%s:%s", u.Scheme, tcp.LOCALHOST, u.Port()))
	m.PushButton(tcp.START)

	m.Sort(kit.SSH_ROUTE)
}

const ROUTE = "route"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ROUTE: {Name: ROUTE, Help: "路由器", Value: kit.Data(kit.MDB_SHORT, kit.SSH_ROUTE)},
		},
		Commands: map[string]*ice.Command{
			ROUTE: {Name: "route route ctx cmd auto invite share", Help: "路由器", Action: map[string]*ice.Action{
				SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					p := kit.MergeURL(m.Option(ice.MSG_USERWEB), SHARE, m.Cmdx(SHARE, mdb.CREATE, kit.MDB_TYPE, LOGIN))
					m.EchoAnchor(p)
					m.EchoScript(p)
					m.EchoQRCode(p)
				}},
				aaa.INVITE: {Name: "invite", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range []string{"tmux", "base", "miss"} {
						m.Cmdy("web.code.publish", "contexts", k)
					}

					m.EchoScript("shell", "# 共享环境", m.Option(ice.MSG_USERWEB))
					m.EchoQRCode(m.Option(ice.MSG_USERWEB))
					m.EchoAnchor(m.Option(ice.MSG_USERWEB))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch m.Option(kit.MDB_ACTION) {
					case mdb.CREATE:
						m.Cmdy(SPACE, m.Option(ROUTE), "web.code.autogen", mdb.INPUTS, arg)
						return
					}

					switch arg[0] {
					case kit.MDB_NAME:
						m.Cmdy(nfs.DIR, m.Conf(DREAM, kit.META_PATH), "name,size,time")
						m.SortStrR(kit.MDB_PATH)

					case kit.MDB_TEMPLATE:
						m.Cmdy(nfs.DIR, m.Conf(DREAM, kit.META_PATH), "path,size,time")
						m.SortStrR(kit.MDB_PATH)
					}
				}},

				mdb.CREATE: {Name: "create main=src/main.go@key name=hi@key from=usr/icebergs/misc/bash/bash.go@key", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option(ROUTE), "web.code.autogen", mdb.CREATE, arg)
					m.ProcessInner()
				}},
				tcp.START: {Name: "start name repos template", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option(ROUTE), DREAM, tcp.START, arg)
					m.ProcessInner()
				}},
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Debug(m.Option(ROUTE))
					m.Cmdy(SPACE, m.Option(ROUTE), kit.Keys(m.Option(kit.SSH_CTX), m.Option(kit.SSH_CMD)), arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || arg[0] == "" { // 路由列表
					if _route_travel(m, kit.Select("", arg, 0)); m.W != nil {
						_route_list(m)
					}

				} else if len(arg) > 2 { // 加载插件
					m.ShowPlugin(arg[0], arg[1], arg[2], kit.MDB_ACTION, ctx.COMMAND)

				} else if len(arg) > 1 { // 命令列表
					m.Cmd(SPACE, arg[0], ctx.CONTEXT, arg[1], ctx.COMMAND).Table(func(index int, value map[string]string, head []string) {
						m.Push(kit.SSH_CMD, value[kit.MDB_KEY])
						m.Push("", value, []string{kit.MDB_NAME, kit.MDB_HELP})
					})

				} else if len(arg) > 0 { // 模块列表
					m.Cmd(SPACE, arg[0], ctx.CONTEXT).Table(func(index int, value map[string]string, head []string) {
						m.Push(kit.SSH_CTX, kit.Keys(value["ups"], value[kit.MDB_NAME]))
						m.Push("", value, []string{ice.CTX_STATUS, ice.CTX_STREAM, kit.MDB_HELP})
					})
				}
			}},
		}})
}
