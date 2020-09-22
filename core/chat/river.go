package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
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
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, USER), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
			m.Push(key, value[kit.MDB_META], []string{kit.MDB_KEY, kit.MDB_NAME}, val[kit.MDB_META])
		})
	})
}

const (
	INFO = "info"
	NODE = "node"
	TOOL = "tool"
	USER = "user"
)
const RIVER = "river"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			RIVER: {Name: RIVER, Help: "群组", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			INFO: {Name: "info auto", Help: "信息", Action: map[string]*ice.Action{
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, RIVER, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, mdb.DETAIL)
				m.Cmdy(mdb.SELECT, RIVER, "", mdb.HASH, kit.MDB_HASH, m.Option(ice.MSG_RIVER))
			}},
			NODE: {Name: "node hash=auto auto 添加 启动", Help: "设备", Action: map[string]*ice.Action{
				"invite": {Name: "invite", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.publish", "contexts", "tool")
				}},
				"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.publish", "contexts", "tool")
				}},
				mdb.INSERT: {Name: "insert route", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH, arg)
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.ROUTE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, "time,hash,route")
				m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH)
				m.Table(func(index int, value map[string]string, head []string) {
					m.Push(kit.MDB_LINK, kit.MergeURL(m.Option(ice.MSG_USERWEB), "pod", kit.Keys(m.Option("pod"), value["route"])))
				})
				m.PushAction("删除")
			}},
			TOOL: {Name: "tool key auto 添加 创建", Help: "工具", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert pod ctx cmd help", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(ice.MSG_STORM)), mdb.LIST, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_ID) != "" {
						m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(ice.MSG_STORM)), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID), arg)
					} else {
						m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
					}
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "pod":
						_storm_list(m, m.Option(ice.MSG_RIVER))
					case "ctx":
						m.Cmdy(ctx.CONTEXT)
					case "cmd", "help":
						m.Cmdy(ctx.CONTEXT, m.Option("ctx"), ctx.COMMAND)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,key,name,count")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH)
				} else {
					m.Option(mdb.FIELDS, "time,id,pod,ctx,cmd,help")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
					if len(m.Appendv(CMD)) == 0 {
						m.Push("time", m.Time())
						m.Push("id", "")
						m.Push("pod", "")
						m.Push("ctx", "")
						m.Push("cmd", arg[1])
						m.Push("help", "")
					}
				}
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
			RIVER: {Name: "river hash auto 添加", Help: "群组", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					h := m.Cmdx(mdb.INSERT, RIVER, "", mdb.HASH, arg)
					m.Option(ice.MSG_RIVER, h)
					m.Echo(h)

					m.Conf(RIVER, kit.Keys(kit.MDB_HASH, h, USER, kit.MDB_META, kit.MDB_SHORT), aaa.USERNAME)
					m.Cmd(USER, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))
					kit.Fetch(m.Confv(RIVER, kit.Keys("meta.template", "base")), func(storm string, value interface{}) {
						h := m.Cmdx(TOOL, mdb.CREATE, kit.MDB_TYPE, "public", kit.MDB_NAME, storm, kit.MDB_TEXT, storm)
						m.Option(ice.MSG_STORM, h)

						kit.Fetch(value, func(index int, value string) {
							m.Search(value, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
								m.Cmd(TOOL, mdb.INSERT, "pod", "", "ctx", s.Cap(ice.CTX_FOLLOW), "cmd", key, "help", kit.Simple(cmd.Help)[0])
							})
						})
					})
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, RIVER, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove hash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, RIVER, "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, RIVER, "", mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(mdb.SELECT, RIVER, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("删除")
			}},

			"/river": {Name: "/river", Help: "小河流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_river_list(m)
					return
				}
				if len(arg) == 1 {
					m.Option(ice.MSG_RIVER, arg[0])
					m.Cmdy(TOOL, arg[1:])
					return
				}
				if !m.Warn(!m.Right(RIVER, arg), ice.ErrNotAuth) {
					m.Cmdy(RIVER, arg)
				}
			}},
		},
	}, nil)
}
