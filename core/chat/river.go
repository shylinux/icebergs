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
func _river_url(m *ice.Message, arg ...string) string {
	return kit.MergeURL(m.Option(ice.MSG_USERWEB), RIVER, m.Option(ice.MSG_RIVER), arg)
}
func _river_list(m *ice.Message) {
	m.Set(ice.MSG_OPTION, kit.MDB_HASH)
	m.Set(ice.MSG_OPTION, kit.MDB_NAME)

	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) {
		case web.RIVER: // 共享群组
			m.Option(ice.MSG_TITLE, msg.Append(kit.MDB_NAME))
			m.Option(ice.MSG_RIVER, msg.Append(RIVER))
			m.Option(ice.MSG_STORM, msg.Append(STORM))

			if m.Conf(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER))) == "" {
				break
			}
			if msg.Cmd(USERS, m.Option(ice.MSG_USERNAME)).Append(aaa.USERNAME) == "" {
				msg.Cmd(USERS, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME)) // 加入群组
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
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, USERS), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
			m.Push(key, kit.GetMeta(value), []string{kit.MDB_HASH, kit.MDB_NAME}, kit.GetMeta(val))
		})
	})
}

const RIVER = "river"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			RIVER: {Name: RIVER, Help: "群组", Value: kit.Data(
				kit.MDB_PATH, ice.USR_LOCAL_RIVER,
				MENUS, `["river",
	["create", "创建群组", "添加应用", "添加工具", "添加设备", "创建空间"],
	["share", "共享群组", "共享应用", "共享工具", "共享主机", "访问空间"]
]`,
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Conf(RIVER, kit.Keym(kit.MDB_TEMPLATE), kit.Dict(
					"base", kit.Dict(
						"info", []interface{}{
							"web.chat.info",
							"web.chat.user",
							"web.chat.tool",
							"web.chat.node",
						},
						"scan", []interface{}{
							"web.chat.scan",
							"web.chat.paste",
							"web.chat.files",
							"web.chat.location",
							"web.chat.meet.miss",
							"web.wiki.feel",
						},
						"task", []interface{}{
							"web.team.task",
							"web.team.plan",
							"web.mall.asset",
							"web.mall.salary",
							"web.wiki.word",
						},
						"draw", []interface{}{
							"web.wiki.draw",
							"web.wiki.data",
							"web.wiki.word",
						},
					),
				))
			}},
			"/river": {Name: "/river", Help: "小河流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
					m.Render(web.STATUS, 401)
					return // 没有登录
				}
				if len(arg) == 0 {
					m.Option(MENUS, m.Conf(RIVER, kit.Keym(MENUS)))
					_river_list(m)
					return // 群组列表
				}
				if len(arg) == 2 && arg[1] == TOOL {
					m.Option(ice.MSG_RIVER, arg[0])
					m.Cmdy(arg[1], arg[2:])
					return // 应用列表
				}
				if m.Warn(!m.Right(RIVER, arg), ice.ErrNotRight) {
					return // 没有授权
				}

				switch kit.Select("", arg, 1) {
				case USERS, TOOL, NODE:
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
				mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello template=base", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					h := m.Cmdx(mdb.INSERT, RIVER, "", mdb.HASH, arg)
					m.Option(ice.MSG_RIVER, h)
					m.Echo(h)

					m.Conf(RIVER, kit.Keys(kit.MDB_HASH, h, NODE, kit.MDB_META, kit.MDB_SHORT), kit.MDB_NAME)
					m.Conf(RIVER, kit.Keys(kit.MDB_HASH, h, USERS, kit.MDB_META, kit.MDB_SHORT), aaa.USERNAME)
					m.Cmd(USERS, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))

					kit.Fetch(m.Confv(RIVER, kit.Keym(kit.MDB_TEMPLATE, kit.Select("base", m.Option(kit.MDB_TEMPLATE)))), func(storm string, value interface{}) {
						h := m.Cmdx(TOOL, mdb.CREATE, kit.MDB_TYPE, PUBLIC, kit.MDB_NAME, storm, kit.MDB_TEXT, storm)

						kit.Fetch(value, func(index int, value string) {
							m.Search(value, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
								m.Cmd(TOOL, mdb.INSERT, kit.MDB_HASH, h, cli.CTX, s.Cap(ice.CTX_FOLLOW), cli.CMD, key, kit.MDB_HELP, cmd.Help)
							})
						})
					})
				}},
				mdb.REMOVE: {Name: "remove", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch m.Option(ctx.ACTION) {
					case cli.START:
						m.Cmdy(web.DREAM, ctx.ACTION, mdb.INPUTS, arg)
						return
					}

					switch arg[0] {
					case aaa.USERNAME:
						m.Cmdy(aaa.USER)
						m.Appendv(ice.MSG_APPEND, aaa.USERNAME, aaa.USERNICK, aaa.USERZONE)
					case aaa.USERROLE:
						m.Push(aaa.USERROLE, aaa.VOID)
						m.Push(aaa.USERROLE, aaa.TECH)
						m.Push(aaa.USERROLE, aaa.ROOT)
					case kit.MDB_TEMPLATE:
						m.Push(kit.MDB_TEMPLATE, "base")
					default:
						m.Cmdy(mdb.INPUTS, RIVER, "", mdb.HASH, arg)
					}
				}},

				aaa.INVITE: {Name: "invite", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.PUBLISH, ice.CONTEXTS)
					m.Cmd(code.PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.BIN_ICE_SH)
					m.Cmd(code.PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.BIN_ICE_BIN)
				}},
				cli.START: {Name: "start name=hi repos template", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(m.Space(m.Option(cli.POD)), web.DREAM, cli.START, arg)
				}},

				SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					_header_share(m, arg...)
				}},
			}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(mdb.SELECT, RIVER, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
