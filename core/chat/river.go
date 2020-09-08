package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _river_list(m *ice.Message) {
	m.Set(ice.MSG_OPTION, kit.MDB_KEY)
	m.Set(ice.MSG_OPTION, kit.MDB_NAME)

	if p := m.Option(POD); p != "" {
		m.Option(POD, "")
		// 代理列表
		m.Cmdy(web.SPACE, p, "web.chat./river")
	}
	m.Richs(RIVER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		if kit.Value(value, "meta.type") == "public" {
			m.Push(key, value[kit.MDB_META], []string{kit.MDB_KEY, kit.MDB_NAME})
			return
		}
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, USER), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
			m.Push(key, value[kit.MDB_META], []string{kit.MDB_KEY, kit.MDB_NAME}, val[kit.MDB_META])
		})
	})
}
func _river_share(m *ice.Message, river, name string, arg ...string) {
	m.Cmdy(web.SHARE, RIVER, name, river, arg)
}
func _river_remove(m *ice.Message, river string) {
	m.Richs(RIVER, nil, river, func(value map[string]interface{}) {
		m.Log_REMOVE(RIVER, river, kit.MDB_VALUE, kit.Format(value))
	})
	m.Conf(RIVER, kit.Keys(kit.MDB_HASH, river), "")
}
func _river_rename(m *ice.Message, river string, name string) {
	prefix := kit.Keys(kit.MDB_HASH, river, kit.MDB_META, kit.MDB_NAME)
	old := m.Conf(RIVER, prefix)
	m.Log_MODIFY(RIVER, river, kit.MDB_VALUE, name, "old", old)
	m.Conf(RIVER, prefix, name)
}
func _river_create(m *ice.Message, kind, name, text string, arg ...string) {
	h := m.Rich(RIVER, nil, kit.Dict(kit.MDB_META, kit.Dict(
		kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
		kit.MDB_EXTRA, kit.Dict(arg),
	),
		USER, kit.Data(kit.MDB_SHORT, aaa.USERNAME),
		TOOL, kit.Data(),
	))
	m.Log_CREATE(kit.MDB_META, RIVER, kit.MDB_TYPE, kind, kit.MDB_NAME, name)

	m.Option(ice.MSG_RIVER, h)
	m.Cmdy(m.Prefix(USER), mdb.INSERT, aaa.USERNAME, cli.UserName)
	m.Cmdy(m.Prefix(USER), mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))

	kit.Fetch(m.Confv(RIVER, kit.Keys("meta.template", "base")), func(storm string, value interface{}) {
		list := []string{}
		kit.Fetch(value, func(index int, value string) {
			m.Search(value, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
				list = append(list, "", s.Cap(ice.CTX_FOLLOW), key, kit.Simple(cmd.Help)[0])
			})
		})
		storm = _storm_create(m, h, "", storm, "")
		_storm_tool(m, h, storm, list...)
	})
	m.Set(ice.MSG_RESULT)
	m.Echo(h)
}

const (
	USER = "user"
	NODE = "node"
	TOOL = "tool"
)
const RIVER = "river"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			RIVER: {Name: "river", Help: "群组", Value: kit.Data(
				"template", kit.Dict(
					"base", kit.Dict(
						"info", []interface{}{
							"web.chat.info",
							"web.chat.tool",
							"web.chat.node",
							"web.chat.user",
						},
						"miss", []interface{}{
							"web.team.task",
							"web.team.plan",
							"web.wiki.word",
						},
					),
				),
				aaa.Black, kit.Dict(aaa.TECH, []interface{}{
					"/river.rename",
					"/river.remove",
					"/storm.remove",
					"/storm.rename",
				}),
				aaa.White, kit.Dict(aaa.VOID, []interface{}{
					"/header",
					"/river",
					"/storm",
					"/action",
					"/footer",
				}),
			)},
		},
		Commands: map[string]*ice.Command{
			"info": {Name: "info auto 导出:button", Help: "信息", Action: map[string]*ice.Action{
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					arg[0] = kit.Keys(kit.MDB_META, arg[0])
					m.Richs(RIVER, nil, m.Option(ice.MSG_RIVER), func(key string, value map[string]interface{}) {
						m.Log_MODIFY(RIVER, m.Option(ice.MSG_RIVER), arg[0], arg[1], "old", kit.Format(kit.Value(value, arg[0])))
						kit.Value(value, arg[0], arg[1])
					})
				}},
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(RIVER), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(RIVER), "", mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(RIVER, nil, m.Option(ice.MSG_RIVER), func(key string, value map[string]interface{}) {
					m.Push("detail", value[kit.MDB_META])
				})
			}},
			TOOL: {Name: "tool hash=auto auto 添加 创建", Help: "工具", Action: map[string]*ice.Action{
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "storm":
						_storm_list(m, m.Option(ice.MSG_RIVER))
					case "ctx":
						m.Cmdy(ctx.COMMAND)
					case "cmd", "help":
						m.Cmdy(ctx.COMMAND)
					}
				}},
				mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_storm_create(m, m.Option(ice.MSG_RIVER), m.Option("type"), m.Option("name"), m.Option("text"))
				}},
				mdb.INSERT: {Name: "insert storm ctx cmd help", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_ID) != "" {
						m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID), arg)
					} else {
						m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,name,count")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH)
				} else {
					m.Option(mdb.FIELDS, "time,id,ctx,cmd,help")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				}
				m.PushAction("删除")
			}},
			NODE: {Name: "node hash=auto auto 添加 启动", Help: "设备", Action: map[string]*ice.Action{
				"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.install", "contexts", "base")
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.ROUTE)
				}},
				mdb.INSERT: {Name: "insert route", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH, arg)
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, "time,hash,route")
				m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH)
				m.Table(func(index int, value map[string]string, head []string) {
					m.Push(kit.MDB_LINK, kit.MergeURL(m.Option(ice.MSG_USERWEB), "pod", kit.Keys(m.Option("pod"), value["route"])))
				})
				m.PushAction("删除")
			}},
			USER: {Name: "user hash=auto auto 添加 邀请", Help: "用户", Action: map[string]*ice.Action{
				"invite": {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.wiki.spark", "inner", kit.MergeURL(m.Option(ice.MSG_USERWEB), "river", m.Option(ice.MSG_RIVER)))
					m.Cmdy("web.wiki.image", "qrcode", kit.MergeURL(m.Option(ice.MSG_USERWEB), "river", m.Option(ice.MSG_RIVER)))
					m.Render("")
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(aaa.USER)
				}},
				mdb.INSERT: {Name: "insert username", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH, arg)
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, "time,hash,username")
				m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH)
				m.Table(func(index int, value map[string]string, head []string) {
					m.Push(aaa.USERZONE, aaa.UserZone(m, value[aaa.USERNAME]))
					m.Push(aaa.USERNICK, aaa.UserNick(m, value[aaa.USERNAME]))
				})
				m.PushAction("删除")
			}},

			"/river": {Name: "/river", Help: "小河流",
				Action: map[string]*ice.Action{
					mdb.CREATE: {Name: "create type name text arg...", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
						_river_create(m, arg[0], arg[1], arg[2], arg[3:]...)
					}},
					mdb.RENAME: {Name: "rename name", Help: "更名", Hand: func(m *ice.Message, arg ...string) {
						_river_rename(m, m.Option(ice.MSG_RIVER), arg[0])
					}},
					mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
						_river_remove(m, m.Option(ice.MSG_RIVER))
					}},
					web.SHARE: {Name: "share name", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
						_river_share(m, m.Option(ice.MSG_RIVER), arg[0])
					}},
				}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
					if len(arg) > 0 && arg[0] == "storm" {
						m.Cmdy("/storm", arg[1:])
						return
					}
					if m.Option("_source") == "" && m.Option(ice.MSG_SESSID) == "" && !tcp.IPIsLocal(m, m.Option(ice.MSG_USERIP)) {
						m.Render("status", "401")
						return
					}

					_river_list(m)
				}},
		},
	}, nil)
}
