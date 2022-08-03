package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _river_key(m *ice.Message, key ...ice.Any) string {
	return kit.Keys(mdb.HASH, m.Option(ice.MSG_RIVER), kit.Simple(key))
}
func _river_list(m *ice.Message) {
	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(mdb.TYPE) {
		case web.RIVER: // 共享群组
			m.Option(ice.MSG_TITLE, msg.Append(mdb.NAME))
			m.Option(ice.MSG_RIVER, msg.Append(RIVER))
			m.Option(ice.MSG_STORM, msg.Append(STORM))

			if m.Conf(RIVER, _river_key(m)) == "" {
				break // 虚拟群组
			}
			if msg.Cmd(OCEAN, m.Option(ice.MSG_USERNAME)).Append(aaa.USERNAME) == "" {
				msg.Cmd(OCEAN, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME)) // 加入群组
			}

		case web.STORM: // 共享应用
			m.Option(ice.MSG_TITLE, msg.Append(mdb.NAME))
			m.Option(ice.MSG_STORM, msg.Append(STORM))
			m.Option(ice.MSG_RIVER, "_share")
			return

		case web.FIELD: // 共享命令
			m.Option(ice.MSG_TITLE, msg.Append(mdb.NAME))
			m.Option(ice.MSG_RIVER, "_share")
			return
		}
	}

	mdb.Richs(m, RIVER, nil, mdb.FOREACH, func(key string, value ice.Map) {
		mdb.Richs(m, RIVER, kit.Keys(mdb.HASH, key, OCEAN), m.Option(ice.MSG_USERNAME), func(k string, val ice.Map) {
			m.Push(key, kit.GetMeta(value), []string{mdb.HASH, mdb.NAME}, kit.GetMeta(val))
		})
	})
}

const RIVER = "river"

func init() {
	Index.MergeCommands(ice.Commands{
		RIVER: {Name: "river hash auto create", Help: "群组", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Config(nfs.TEMPLATE, _river_template) }},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case cli.START, "创建空间":
					m.Cmdy(web.DREAM, mdb.INPUTS, arg)
					return
				}

				switch arg[0] {
				case nfs.TEMPLATE:
					m.Push(nfs.TEMPLATE, ice.BASE)
				case aaa.USERROLE:
					m.Push(aaa.USERROLE, aaa.VOID, aaa.TECH, aaa.ROOT)
				case aaa.USERNAME:
					m.Cmdy(aaa.USER).Cut(aaa.USERNAME, aaa.USERNICK, aaa.USERZONE)
				default:
					m.Cmdy(mdb.INPUTS, RIVER, "", mdb.HASH, arg)
				}
			}},
			mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello template=base", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				h := m.Cmdx(mdb.INSERT, RIVER, "", mdb.HASH, arg)
				m.Option(ice.MSG_RIVER, h)
				m.Echo(h)

				m.Conf(RIVER, kit.Keys(mdb.HASH, h, NODE, kit.Keym(mdb.SHORT)), mdb.NAME)
				m.Conf(RIVER, kit.Keys(mdb.HASH, h, OCEAN, kit.Keym(mdb.SHORT)), aaa.USERNAME)
				m.Cmd(OCEAN, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))

				kit.Fetch(m.Confv(RIVER, kit.Keym(nfs.TEMPLATE, kit.Select("base", m.Option(nfs.TEMPLATE)))), func(storm string, value ice.Any) {
					h := m.Cmdx(STORM, mdb.CREATE, mdb.TYPE, PUBLIC, mdb.NAME, storm, mdb.TEXT, storm)

					kit.Fetch(value, func(index int, value string) {
						m.Search(value, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
							m.Cmd(STORM, mdb.INSERT, mdb.HASH, h, ice.CTX, s.Cap(ice.CTX_FOLLOW), ice.CMD, key, mdb.HELP, cmd.Help)
						})
					})
				})
			}},
			cli.START: {Name: "start name=hi repos template", Help: "创建空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.Space(m, m.Option(ice.POD)), web.DREAM, cli.START, arg)
			}},
			aaa.INVITE: {Name: "invite", Help: "添加设备", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(code.PUBLISH, mdb.CREATE, nfs.FILE, ice.BIN_ICE_BIN)
				m.Cmdy(code.PUBLISH, ice.CONTEXTS)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,template"))},
		web.P(RIVER): {Name: "/river", Help: "群组", Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, RIVER) {
				m.RenderStatusUnauthorized()
				return // 没有登录
			}
			if len(arg) == 0 {
				m.Option(MENUS, m.Config(MENUS))
				_river_list(m)
				return // 群组列表
			}
			if len(arg) == 2 && arg[1] == STORM {
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy(arg[1], arg[2:])
				return // 应用列表
			}
			if !aaa.Right(m, RIVER, arg) {
				m.RenderStatusForbidden()
				return // 没有授权
			}

			switch kit.Select("", arg, 1) {
			case STORM, OCEAN, NODE:
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy(arg[1], arg[2:])

			case ctx.ACTION, aaa.INVITE:
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy(RIVER, arg[1:])

			default:
				m.Cmdy(RIVER, arg)
			}
		}},
	})
}

var _river_template = kit.Dict(
	"base", kit.Dict(
		"info", kit.List(
			"web.chat.info",
			"web.chat.ocean",
			"web.chat.storm",
			"web.chat.node",
		),
		"scan", kit.List(
			"web.chat.scan",
			"web.chat.paste",
			"web.chat.files",
			"web.chat.location",
			"web.chat.meet.miss",
			"web.wiki.feel",
		),
		"task", kit.List(
			"web.team.task",
			"web.team.plan",
			"web.mall.asset",
			"web.mall.salary",
			"web.wiki.word",
		),
		"draw", kit.List(
			"web.wiki.draw",
			"web.wiki.data",
			"web.wiki.word",
		),
	),
)
