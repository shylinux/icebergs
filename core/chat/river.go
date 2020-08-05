package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
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
func _river_node(m *ice.Message, river string, node ...string) {
	prefix := kit.Keys(kit.MDB_HASH, river, NODE)
	for _, v := range node {
		m.Rich(RIVER, prefix, kit.Data(kit.MDB_NAME, v))
		m.Log_INSERT(RIVER, river, NODE, v)
	}
}
func _river_user(m *ice.Message, river string, user ...string) {
	prefix := kit.Keys(kit.MDB_HASH, river, USER)
	for _, v := range user {
		m.Rich(RIVER, prefix, kit.Data(aaa.USERNAME, v))
		m.Log_INSERT(RIVER, river, USER, v)
	}
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

	_river_user(m, h, cli.UserName, m.Option(ice.MSG_USERNAME))
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
						"admin", []interface{}{
							"web.chat.user",
							"web.chat.tool",
							"web.chat.info",
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

			"user": {Name: "user auto", Help: "用户", Action: map[string]*ice.Action{
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(aaa.USERNAME) == cli.UserName {
						return
					}
					if m.Option(aaa.USERNAME) == m.Option(ice.MSG_USERNAME) {
						return
					}
					m.Richs(RIVER, nil, m.Option(ice.MSG_RIVER), func(key string, value map[string]interface{}) {
						m.Richs(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), m.Option(aaa.USERNAME), func(key string, value map[string]interface{}) {
							m.Conf(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER, kit.MDB_HASH, key), "")
							m.Log_DELETE(RIVER, m.Option(ice.MSG_RIVER), USER, m.Option(aaa.USERNAME))
						})
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(RIVER, nil, m.Option(ice.MSG_RIVER), func(key string, value map[string]interface{}) {
					m.Richs(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						value = value[kit.MDB_META].(map[string]interface{})
						m.Push(key, value, []string{kit.MDB_TIME})
						m.Push(aaa.USERZONE, aaa.UserZone(m, value[aaa.USERNAME]))
						m.Push(aaa.USERNICK, aaa.UserNick(m, value[aaa.USERNAME]))
						m.Push(key, value, []string{aaa.USERNAME})
					})
				})
				m.PushAction("删除")
			}},
			"tool": {Name: "tool storm=auto id=auto auto", Help: "工具", Action: map[string]*ice.Action{
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(RIVER, nil, m.Option(ice.MSG_RIVER), func(key string, value map[string]interface{}) {
						m.Richs(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), m.Option("storm"), func(key string, value map[string]interface{}) {
							switch arg[0] {
							case kit.MDB_TIME, kit.MDB_ID:
								return
							}
							if m.Option(kit.MDB_ID) == "" {
								m.Log_MODIFY(RIVER, m.Option(ice.MSG_RIVER), arg[0], arg[1], "old", kit.Value(value, kit.Keys(kit.MDB_META, arg[0])))
								kit.Value(value, kit.Keys(kit.MDB_META, arg[0]), arg[1])
								return
							}
							m.Grows(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, key), kit.MDB_ID, m.Option(kit.MDB_ID), func(index int, value map[string]interface{}) {
								m.Log_MODIFY(RIVER, m.Option(ice.MSG_RIVER), arg[0], arg[1], "old", kit.Value(value, kit.Keys(kit.MDB_META, arg[0])))
								kit.Value(value, kit.Keys(kit.MDB_META, arg[0]), arg[1])
							})
						})
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Richs(RIVER, nil, m.Option(ice.MSG_RIVER), func(key string, value map[string]interface{}) {
						m.Richs(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
							value = value[kit.MDB_META].(map[string]interface{})
							m.Push("time", value[kit.MDB_TIME])
							m.Push("storm", key)
							m.Push("name", value[kit.MDB_NAME])
							m.Push("count", value[kit.MDB_COUNT])
						})
					})
					return
				}

				m.Richs(RIVER, nil, m.Option(ice.MSG_RIVER), func(key string, value map[string]interface{}) {
					if len(arg) == 1 {
						m.Grows(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, arg[0]), "", "", func(index int, value map[string]interface{}) {
							value = value[kit.MDB_META].(map[string]interface{})
							m.Push("time", value[kit.MDB_TIME])
							m.Push("id", value[kit.MDB_ID])
							m.Push("ctx", value["ctx"])
							m.Push("cmd", value["cmd"])
						})
						return
					}
					m.Grows(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, arg[0]), kit.MDB_ID, arg[1], func(index int, value map[string]interface{}) {
						value = value[kit.MDB_META].(map[string]interface{})
						m.Push("detail", value)
					})
				})
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
					USER: {Name: "user user...", Help: "添加用户", Hand: func(m *ice.Message, arg ...string) {
						_river_user(m, m.Option(ice.MSG_RIVER), arg...)
					}},
					NODE: {Name: "node node...", Help: "添加设备", Hand: func(m *ice.Message, arg ...string) {
						_river_node(m, m.Option(ice.MSG_RIVER), arg...)
					}},
				}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
					if m.Option(ice.MSG_USERNICK) != "" {
						m.Cmd(aaa.USER, mdb.MODIFY, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK))
					}
					if len(arg) > 0 && arg[0] == "storm" {
						m.Cmdy("/storm", arg[1:])
						return
					}
					if m.Option("_source") == "" && m.Option(ice.MSG_SESSID) == "" {
						m.Render("status", "401")
						return
					}

					_river_list(m)
				}},
		},
	}, nil)
}
