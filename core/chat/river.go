package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _river_key(m *ice.Message, key ...interface{}) string {
	return kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), kit.Simple(key))
}
func _river_list(m *ice.Message) {
	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) {
		case web.RIVER: // 共享群组
			m.Option(ice.MSG_TITLE, msg.Append(kit.MDB_NAME))
			m.Option(ice.MSG_RIVER, msg.Append(RIVER))
			m.Option(ice.MSG_STORM, msg.Append(STORM))

			if m.Conf(RIVER, _river_key(m)) == "" {
				break // 虚拟群组
			}
			if msg.Cmd(OCEAN, m.Option(ice.MSG_USERNAME)).Append(aaa.USERNAME) == "" {
				msg.Cmd(OCEAN, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME)) // 加入群组
			}

		case web.STORM: // 共享应用
			m.Option(ice.MSG_TITLE, msg.Append(kit.MDB_NAME))
			m.Option(ice.MSG_STORM, msg.Append(STORM))
			m.Option(ice.MSG_RIVER, "_share")
			return

		case web.FIELD: // 共享命令
			m.Option(ice.MSG_TITLE, msg.Append(kit.MDB_NAME))
			m.Option(ice.MSG_RIVER, "_share")
			return
		}
	}

	m.Richs(RIVER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, OCEAN), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
			m.Push(key, kit.GetMeta(value), []string{kit.MDB_HASH, kit.MDB_NAME}, kit.GetMeta(val))
		})
	})
}

const RIVER = "river"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		RIVER: {Name: RIVER, Help: "群组", Value: kit.Data(
			MENUS, kit.List(RIVER, kit.List("create", "创建群组", "添加应用", "添加工具", "添加设备", "创建空间"), kit.List("share", "共享群组", "共享应用", "共享工具", "共享主机", "访问空间")),
		)},
	}, Commands: map[string]*ice.Command{
		"/river": {Name: "/river", Help: "小河流", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Config(kit.MDB_TEMPLATE, kit.Dict(
					"base", kit.Dict(
						"info", kit.List(
							"web.chat.info",
							"web.chat.user",
							"web.chat.tool",
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
				))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, RIVER) {
				m.Render(web.STATUS, 401)
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
			if !m.Right(RIVER, arg) {
				return // 没有授权
			}

			switch kit.Select("", arg, 1) {
			case OCEAN, NODE:
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy(arg[1], arg[2:])

			case ctx.ACTION, aaa.INVITE:
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy(RIVER, arg[1:])

			default:
				m.Cmdy(RIVER, arg)
			}
		}},
		RIVER: {Name: "river hash auto create", Help: "群组", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case cli.START:
					m.Cmdy(web.DREAM, ctx.ACTION, mdb.INPUTS, arg)
					return
				}

				switch arg[0] {
				case kit.MDB_TEMPLATE:
					m.Push(kit.MDB_TEMPLATE, ice.BASE)
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

				m.Conf(RIVER, kit.Keys(kit.MDB_HASH, h, NODE, kit.Keym(kit.MDB_SHORT)), kit.MDB_NAME)
				m.Conf(RIVER, kit.Keys(kit.MDB_HASH, h, OCEAN, kit.Keym(kit.MDB_SHORT)), aaa.USERNAME)
				m.Cmd(OCEAN, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))

				kit.Fetch(m.Confv(RIVER, kit.Keym(kit.MDB_TEMPLATE, kit.Select("base", m.Option(kit.MDB_TEMPLATE)))), func(storm string, value interface{}) {
					h := m.Cmdx(STORM, mdb.CREATE, kit.MDB_TYPE, PUBLIC, kit.MDB_NAME, storm, kit.MDB_TEXT, storm)

					kit.Fetch(value, func(index int, value string) {
						m.Search(value, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
							m.Cmd(STORM, mdb.INSERT, kit.MDB_HASH, h, ice.CTX, s.Cap(ice.CTX_FOLLOW), ice.CMD, key, kit.MDB_HELP, cmd.Help)
						})
					})
				})
			}},

			cli.START: {Name: "start name=hi repos template", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(m.Space(m.Option(ice.POD)), web.DREAM, cli.START, arg)
			}},
			aaa.INVITE: {Name: "invite", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.PUBLISH, ice.CONTEXTS)
				m.Cmd(code.PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.BIN_ICE_SH)
				m.Cmd(code.PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.BIN_ICE_BIN)
			}},
			SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_header_share(m, arg...)
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
