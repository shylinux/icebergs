package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"
)

func _river_list(m *ice.Message) {
	if m.Option("share") != "" && m.Option("share") != "" {
		m.Option(ice.MSG_RIVER, m.Option("river"))
		if m.Cmd(AUTH, m.Option("share")).Append(kit.MDB_TYPE) == USER {
			if m.Cmd(m.Prefix(USER), m.Option(ice.MSG_USERNAME)).Append(ice.MSG_USERNAME) == "" {
				m.Cmd(m.Prefix(USER), mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))
				// 加入群组
			}
		}
	}

	m.Set(ice.MSG_OPTION, kit.MDB_KEY)
	m.Set(ice.MSG_OPTION, kit.MDB_NAME)

	if p := m.Option(POD); p != "" {
		m.Option(POD, "")
		m.Cmdy(web.SPACE, p, "web.chat./river")
		// 代理列表
	}

	m.Richs(RIVER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, USER), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
			m.Push(key, value[kit.MDB_META], []string{kit.MDB_HASH, kit.MDB_NAME}, val[kit.MDB_META])
		})
	})
}
func _river_proxy(m *ice.Message, pod string) (proxy []string) {
	if p := kit.Select(m.Option(POD), pod); p != "" {
		proxy = append(proxy, web.SPACE, p)
		m.Option(POD, "")
	}
	return proxy
}

const (
	POD = "pod"
	CTX = "ctx"
	CMD = "cmd"
)
const (
	INFO = "info"
	AUTH = "auth"
	NODE = "node"
	TOOL = "tool"
	USER = "user"
)
const STORM = "storm"
const RIVER = "river"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			RIVER: {Name: RIVER, Help: "群组", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			INFO: {Name: "info auto", Help: "信息", Action: map[string]*ice.Action{
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, RIVER, "", mdb.HASH, kit.MDB_HASH, m.Option(ice.MSG_RIVER), arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, mdb.DETAIL)
				m.Cmdy(mdb.SELECT, RIVER, "", mdb.HASH, kit.MDB_HASH, m.Option(ice.MSG_RIVER))
			}},
			AUTH: {Name: "auth hash auto 添加", Help: "授权", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=node,user name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), AUTH), mdb.HASH,
						aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME),
						kit.MDB_TIME, m.Time("72h"), arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), AUTH), mdb.HASH,
						kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), AUTH), mdb.HASH,
						kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, "time,hash,userrole,username,type,name,text")
				m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), AUTH), mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("删除")
			}},
			NODE: {Name: "node name ctx cmd auto 邀请", Help: "设备", Action: map[string]*ice.Action{
				mdb.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					m.Option(web.SHARE, m.Cmdx(AUTH, mdb.CREATE, kit.MDB_TYPE, NODE))
					m.Cmdy(code.PUBLISH, "contexts", "tool")
				}},
				mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(RIVER), NODE), mdb.HASH, arg)
				}},
				web.SPACE_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
						return
					}
					if msg := m.Cmd(AUTH, m.Option("share")); msg.Append(kit.MDB_TYPE) == NODE {
						m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(RIVER), NODE), mdb.HASH, arg)
					}
				}},
				web.SPACE_CLOSE: {Name: "close", Help: "关闭", Hand: func(m *ice.Message, arg ...string) {
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(RIVER), NODE), mdb.HASH,
						kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,type,name,share")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH)
					m.PushAction("删除")
					return
				}
				m.Cmdy(web.ROUTE, arg)
			}},
			TOOL: {Name: "tool hash id auto 添加 创建", Help: "工具", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert hash pod ctx cmd help", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST, arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_ID) != "" {
						m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST,
							kit.MDB_ID, m.Option(kit.MDB_ID), arg)
					} else {
						m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH,
							kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
					}
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH,
						kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_HASH:
						m.Cmd(TOOL).Table(func(index int, value map[string]string, head []string) {
							m.Push(kit.MDB_HASH, value[kit.MDB_HASH])
							m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
						})
					case POD:
						m.Cmdy(web.ROUTE)
					case CTX:
						m.Cmd(ctx.CONTEXT, "web").Table(func(index int, value map[string]string, head []string) {
							m.Push(CTX, kit.Keys(kit.Select("", value["ups"], value["ups"] != "shy"), value[kit.MDB_NAME]))
							m.Push(kit.MDB_HELP, value[kit.MDB_HELP])
						})
					case CMD, "help":
						m.Cmd(ctx.CONTEXT, m.Option(CTX), ctx.COMMAND).Table(func(index int, value map[string]string, head []string) {
							m.Push(CMD, value[kit.MDB_KEY])
							m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
							m.Push(kit.MDB_HELP, value[kit.MDB_HELP])
						})
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,name,count")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH)
					m.PushAction("删除")
					return
				}

				m.Option(mdb.FIELDS, "time,id,pod,ctx,cmd,arg")
				msg := m.Cmd(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, kit.Select("", arg, 1))
				if len(msg.Appendv(CMD)) == 0 && len(arg) > 1 {
					msg.Push(CMD, arg[1])
				}

				if len(arg) > 2 && arg[2] == "run" {
					m.Cmdy(_river_proxy(m, msg.Append(POD)), kit.Keys(msg.Append(CTX), msg.Append(CMD)), arg[3:])
					return
				}
				if m.Copy(msg); len(arg) < 2 {
					return
				}

				m.Option("_process", "_field")
				m.Option("_prefix", arg[0], arg[1], "run")
				m.Table(func(index int, value map[string]string, head []string) {
					m.Cmdy(web.SPACE, value[POD], ctx.CONTEXT, value[CTX], ctx.COMMAND, value[CMD])
				})
			}},
			USER: {Name: "user username auto 邀请", Help: "用户", Action: map[string]*ice.Action{
				mdb.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					share := m.Option(web.SHARE, m.Cmdx(AUTH, mdb.CREATE, kit.MDB_TYPE, USER))
					m.Cmdy(wiki.SPARK, "inner", kit.MergeURL(m.Option(ice.MSG_USERWEB), "river", m.Option(ice.MSG_RIVER), "share", share))
					m.Cmdy(wiki.IMAGE, "qrcode", kit.MergeURL(m.Option(ice.MSG_USERWEB), "river", m.Option(ice.MSG_RIVER), "share", share))
					m.Render("")
				}},
				mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH, arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH, aaa.USERNAME, m.Option(aaa.USERNAME))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, "time,username")
				m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH, aaa.USERNAME, arg)
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
					m.Conf(RIVER, kit.Keys(kit.MDB_HASH, h, NODE, kit.MDB_META, kit.MDB_SHORT), kit.MDB_NAME)
					m.Cmd(m.Prefix(USER), mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))

					kit.Fetch(m.Confv(RIVER, kit.Keys("meta.template", "base")), func(storm string, value interface{}) {
						h := m.Cmdx(TOOL, mdb.CREATE, kit.MDB_TYPE, "public", kit.MDB_NAME, storm, kit.MDB_TEXT, storm)

						kit.Fetch(value, func(index int, value string) {
							m.Search(value, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
								m.Cmd(TOOL, mdb.INSERT, kit.MDB_HASH, h, CTX, s.Cap(ice.CTX_FOLLOW), CMD, key, kit.MDB_HELP, kit.Simple(cmd.Help)[0])
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
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, RIVER, "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(mdb.SELECT, RIVER, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("删除")
			}},

			"/river": {Name: "/river", Help: "小河流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
					return
				}
				if len(arg) == 0 {
					_river_list(m)
					return
				}
				switch kit.Select("", arg, 1) {
				case USER, TOOL, NODE:
					m.Option(ice.MSG_RIVER, arg[0])
					m.Cmdy(m.Prefix(arg[1]), arg[2:])
					return
				}
				if !m.Warn(!m.Right(RIVER, arg), ice.ErrNotAuth) {
					m.Cmdy(RIVER, arg)
				}
			}},
		},
	}, nil)
}
