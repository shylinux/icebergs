package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _route_travel(m *ice.Message, route string) {
	m.Cmd(SPACE).Tables(func(val ice.Maps) {
		switch val[mdb.TYPE] {
		case SERVER: // 远程查询
			if val[mdb.NAME] == ice.Info.NodeName {
				return // 避免循环
			}

			m.Cmd(SPACE, val[mdb.NAME], ROUTE).Tables(func(value ice.Maps) {
				m.Push(mdb.TYPE, value[mdb.TYPE])
				m.Push(ROUTE, kit.Keys(val[mdb.NAME], value[ROUTE]))
			})
			fallthrough
		case WORKER: // 本机查询
			m.Push(mdb.TYPE, val[mdb.TYPE])
			m.Push(ROUTE, val[mdb.NAME])
		}
	})
}
func _route_list(m *ice.Message) *ice.Message {
	m.Tables(func(value ice.Maps) {
		switch m.PushAnchor(value[ROUTE], MergePod(m, value[ROUTE])); value[mdb.TYPE] {
		case SERVER:
			m.PushButton(tcp.START, aaa.INVITE)
		case WORKER:
			fallthrough
		default:
			m.PushButton("")
		}
	})

	// 网卡信息
	u := OptionUserWeb(m)
	m.Cmd(tcp.HOST).Tables(func(value ice.Maps) {
		m.Push(mdb.TYPE, MYSELF)
		m.Push(ROUTE, ice.Info.NodeName)
		m.PushAnchor(value[aaa.IP], kit.Format("%s://%s:%s", u.Scheme, value[aaa.IP], u.Port()))
		m.PushButton(tcp.START)
	})

	// 本机信息
	m.Push(mdb.TYPE, MYSELF)
	m.Push(ROUTE, ice.Info.NodeName)
	m.PushAnchor(tcp.LOCALHOST, kit.Format("%s://%s:%s", u.Scheme, tcp.LOCALHOST, u.Port()))
	m.PushButton(tcp.START)
	return m
}

const ROUTE = "route"

func init() {
	Index.MergeCommands(ice.Commands{
		ROUTE: {Name: "route route ctx cmd auto invite spide", Help: "路由器", Actions: ice.Actions{
			aaa.INVITE: {Name: "invite", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPACE, m.Option(ROUTE), SPACE, aaa.INVITE, arg).ProcessInner()
			}},
			cli.START: {Name: "start name repos template", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPACE, m.Option(ROUTE), DREAM, tcp.START, arg).ProcessInner()
			}},
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPACE, m.Option(ROUTE), kit.Keys(m.Option(ice.CTX), m.Option(ice.CMD)), arg)
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPACE, m.Option(ROUTE), arg)
			}},
			SPIDE: {Name: "spide", Help: "架构图", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ROUTE) == "" { // 路由列表 route
					ctx.DisplayStorySpide(m, lex.PREFIX, SPIDE, lex.SPLIT, ice.PT)
					m.Cmdy(ROUTE).Cut(ROUTE)

				} else if m.Option(ctx.CONTEXT) == "" { // 模块列表 context
					m.Cmdy(SPACE, m.Option(ROUTE), ctx.CONTEXT, ice.ICE, ctx.CONTEXT).Cut(mdb.NAME).RenameAppend(mdb.NAME, ctx.CONTEXT)
					m.Option(lex.SPLIT, ice.PT)

				} else if m.Option(mdb.NAME) == "" { // 命令列表 name
					m.Cmdy(SPACE, m.Option(ROUTE), ctx.CONTEXT, SPIDE, "", m.Option(ctx.CONTEXT), m.Option(ctx.CONTEXT)).Cut(mdb.NAME)

				} else { // 命令详情 index name help meta list
					m.Cmdy(SPACE, m.Option(ROUTE), ctx.CONTEXT, SPIDE, "", m.Option(ctx.CONTEXT), m.Option(mdb.NAME))
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] == "" { // 路由列表
				if _route_travel(m, kit.Select("", arg, 0)); m.W != nil {
					_route_list(m).Sort("type,route")
				}

			} else if len(arg) == 1 || arg[1] == "" { // 模块列表
				m.Cmd(SPACE, arg[0], ctx.CONTEXT, ice.ICE).Tables(func(value ice.Maps) {
					m.Push(ice.CTX, kit.Keys(value["ups"], value[mdb.NAME]))
					m.Push("", value, kit.List(ice.CTX_STATUS, ice.CTX_STREAM, mdb.HELP))
				})

			} else if len(arg) == 2 || arg[2] == "" { // 命令列表
				m.Cmd(SPACE, arg[0], ctx.CONTEXT, arg[1], ctx.COMMAND).Tables(func(value ice.Maps) {
					m.Push(ice.CMD, value[mdb.KEY])
					m.Push("", value, kit.List(mdb.NAME, mdb.HELP))
				})

			} else { // 命令详情
				m.Cmdy(SPACE, arg[0], ctx.CONTEXT, arg[1], ctx.COMMAND, arg[2])
				m.ProcessField(ctx.ACTION, ctx.COMMAND)
			}
		}},
	})
}
