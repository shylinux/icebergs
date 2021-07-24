package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

func _river_list(m *ice.Message) {
	m.Set(ice.MSG_OPTION, kit.MDB_HASH)
	m.Set(ice.MSG_OPTION, kit.MDB_NAME)

	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) {
		case web.RIVER: // 应用入口
			m.Option(ice.MSG_TITLE, msg.Append(kit.MDB_NAME))
			m.Option(ice.MSG_RIVER, msg.Append(RIVER))
			m.Option(ice.MSG_STORM, msg.Append(STORM))

			if m.Conf(RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER))) == "" {
				break
			}
			if msg.Cmd(m.Prefix(USER), m.Option(ice.MSG_USERNAME)).Append(aaa.USERNAME) == "" {
				msg.Cmd(m.Prefix(USER), mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))
				// 加入群组
			}
		case web.STORM: // 应用入口
			m.Option(ice.MSG_TITLE, msg.Append(kit.MDB_NAME))
			m.Option(ice.MSG_STORM, msg.Append(STORM))
			m.Option(ice.MSG_RIVER, "_share")

		case web.FIELD: // 应用入口
			m.Option(ice.MSG_TITLE, msg.Append(kit.MDB_NAME))
			m.Option(ice.MSG_RIVER, "_share")
			return
		}
	}

	m.Richs(RIVER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, USER), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
			m.Push(key, kit.GetMeta(value), []string{kit.MDB_HASH, kit.MDB_NAME}, kit.GetMeta(val))
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
	ARG = "arg"
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
			RIVER: {Name: RIVER, Help: "群组", Value: kit.Data(kit.MDB_PATH, "usr/local/river")},
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
			AUTH: {Name: "auth hash auto create", Help: "授权", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=node,user name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), AUTH), mdb.HASH,
						aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME),
						kit.MDB_TIME, m.Time("72h"), arg)
				}},
				mdb.INSERT: {Name: "insert river share", Help: "加入", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,userrole,username,type,name,text")
					msg := m.Cmd(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(RIVER), AUTH), mdb.HASH, kit.MDB_HASH, m.Option(web.SHARE))
					switch msg.Append(kit.MDB_TYPE) {
					case USER:
						m.Option(ice.MSG_RIVER, m.Option(RIVER))
						m.Cmdy(m.Prefix(USER), mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))
					case NODE:
					}
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
				m.Option(mdb.FIELDS, kit.Select("time,hash,userrole,username,type,name,text", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), AUTH), mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)
				if len(arg) > 0 {
					m.PushQRCode("qrcode", kit.MergeURL(m.Option(ice.MSG_USERWEB), RIVER, m.Option(ice.MSG_RIVER), web.SHARE, m.Option("share")))
					m.PushScript("script", kit.MergeURL(m.Option(ice.MSG_USERWEB), RIVER, m.Option(ice.MSG_RIVER), web.SHARE, m.Option("share")))
				}
			}},
			NODE: {Name: "node name ctx cmd auto insert invite", Help: "设备", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert type name share", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(RIVER), NODE), mdb.HASH, arg)
				}},
				aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					m.Option(web.SHARE, m.Cmdx(m.Prefix(AUTH), mdb.CREATE, kit.MDB_TYPE, NODE))
					m.Cmdy(code.PUBLISH, "contexts", "tool")
				}},
				web.SPACE_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
						return
					}
					if msg := m.Cmd(m.Prefix(AUTH), m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) == NODE {
						m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(RIVER), NODE), mdb.HASH, arg)
					}
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(RIVER), NODE), mdb.HASH,
						kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.SPACE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,type,name,share")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH)
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushAnchor(value[kit.MDB_NAME], kit.MergeURL(m.Option(ice.MSG_USERWEB),
							kit.SSH_POD, kit.Keys(m.Option(kit.SSH_POD), value[kit.MDB_NAME])))
					})
					m.PushAction(mdb.REMOVE)
					return
				}
				m.Cmdy(web.ROUTE, arg)
			}},
			TOOL: {Name: "tool hash id auto insert create", Help: "工具", Action: map[string]*ice.Action{
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
						m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
					}
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH,
						kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_ID) != "" {
						m.Option(mdb.FIELDS, "time,id,pod,ctx,cmd,arg")
						msg := m.Cmd(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID))

						cmd := kit.Keys(msg.Append(CTX), msg.Append(CMD))
						_action_domain(m, cmd, m.Option(kit.MDB_HASH))
						m.Cmdy(_river_proxy(msg, msg.Append(POD)), cmd, mdb.EXPORT)
					}
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_ID) != "" {
						m.Option(mdb.FIELDS, "time,id,pod,ctx,cmd,arg")
						msg := m.Cmd(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID))

						cmd := kit.Keys(msg.Append(CTX), msg.Append(CMD))
						_action_domain(m, cmd, m.Option(kit.MDB_HASH))
						m.Cmdy(_river_proxy(msg, msg.Append(POD)), cmd, mdb.IMPORT)
					}
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
					case CMD, kit.MDB_HELP:
						m.Cmd(ctx.CONTEXT, m.Option(CTX), ctx.COMMAND).Table(func(index int, value map[string]string, head []string) {
							m.Push(CMD, value[kit.MDB_KEY])
							m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
							m.Push(kit.MDB_HELP, value[kit.MDB_HELP])
						})
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,type,name,count")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL), mdb.HASH)
					m.PushAction(mdb.REMOVE)
					m.Sort(kit.MDB_NAME)
					return // 应用列表
				}

				m.Option(mdb.FIELDS, "time,id,pod,ctx,cmd,arg,display,style")
				msg := m.Cmd(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), TOOL, kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, kit.Select("", arg, 1))
				if len(msg.Appendv(CMD)) == 0 && len(arg) > 1 {
					msg.Push(CMD, arg[1])
				}

				if len(arg) > 2 && arg[2] == "run" {
					m.Cmdy(_river_proxy(m, msg.Append(POD)), kit.Keys(msg.Append(CTX), msg.Append(CMD)), arg[3:])
					return // 执行命令
				}
				if m.Copy(msg); len(arg) < 2 {
					m.PushAction(mdb.EXPORT, mdb.IMPORT)
					m.SortInt(kit.MDB_ID)
					return // 命令列表
				}

				// 命令插件
				m.ProcessField(arg[0], arg[1], "run")
				m.Table(func(index int, value map[string]string, head []string) {
					m.Cmdy(web.SPACE, value[POD], ctx.CONTEXT, value[CTX], ctx.COMMAND, value[CMD])
				})
			}},
			USER: {Name: "user username auto insert invite", Help: "用户", Action: map[string]*ice.Action{
				aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					share := m.Option(web.SHARE, m.Cmdx(m.Prefix(AUTH), mdb.CREATE, kit.MDB_TYPE, USER))
					m.EchoScript(kit.MergeURL(m.Option(ice.MSG_USERWEB), RIVER, m.Option(ice.MSG_RIVER), web.SHARE, share))
					m.EchoQRCode(kit.MergeURL(m.Option(ice.MSG_USERWEB), RIVER, m.Option(ice.MSG_RIVER), web.SHARE, share))
					m.Render("")
				}},
				mdb.INSERT: {Name: "insert username", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH, arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH, aaa.USERNAME, m.Option(aaa.USERNAME))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("aaa.user")
					m.Appendv(ice.MSG_APPEND, aaa.USERNAME, aaa.USERZONE, aaa.USERNICK)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,username", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), USER), mdb.HASH, aaa.USERNAME, arg)
				m.Table(func(index int, value map[string]string, head []string) {
					m.Richs(USER, nil, value[aaa.USERNAME], func(key string, val map[string]interface{}) {
						val = kit.GetMeta(val)
						m.Push(aaa.USERNICK, val[aaa.USERNICK])
						m.PushImages(aaa.AVATAR, kit.Format(val[aaa.AVATAR]), kit.Select("60", "240", m.Option(mdb.FIELDS) == mdb.DETAIL))
					})
				})
				m.PushAction(mdb.REMOVE)
			}},
			RIVER: {Name: "river hash auto create", Help: "群组", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello template=base", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					h := m.Cmdx(mdb.INSERT, RIVER, "", mdb.HASH, arg)
					m.Option(ice.MSG_RIVER, h)
					m.Echo(h)

					m.Conf(RIVER, kit.Keys(kit.MDB_HASH, h, NODE, kit.MDB_META, kit.MDB_SHORT), kit.MDB_NAME)
					m.Conf(RIVER, kit.Keys(kit.MDB_HASH, h, USER, kit.MDB_META, kit.MDB_SHORT), aaa.USERNAME)
					m.Cmd(m.Prefix(USER), mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))

					kit.Fetch(m.Confv(RIVER, kit.Keym(kit.MDB_TEMPLATE, kit.Select("base", m.Option(kit.MDB_TEMPLATE)))), func(storm string, value interface{}) {
						h := m.Cmdx(TOOL, mdb.CREATE, kit.MDB_TYPE, PUBLIC, kit.MDB_NAME, storm, kit.MDB_TEXT, storm)

						kit.Fetch(value, func(index int, value string) {
							m.Search(value, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
								m.Cmd(TOOL, mdb.INSERT, kit.MDB_HASH, h, CTX, s.Cap(ice.CTX_FOLLOW), CMD, key, kit.MDB_HELP, cmd.Help)
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
				aaa.INVITE: {Name: "invite", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range []string{"tmux", "base", "miss", "binary", "source", "module"} {
						m.Cmdy("web.code.publish", "contexts", k)
					}
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch m.Option(kit.MDB_ACTION) {
					case tcp.START:
						m.Cmdy(web.DREAM, kit.MDB_ACTION, mdb.INPUTS, arg)
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
				tcp.START: {Name: "start name repos template", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(m.Space(m.Option(kit.SSH_POD)), web.DREAM, tcp.START, arg)
				}},

				SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					_header_share(m, arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(mdb.SELECT, RIVER, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)
			}},

			"/river": {Name: "/river", Help: "小河流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
					m.Render(web.STATUS, 401)
					return // 没有登录
				}
				if len(arg) == 0 {
					_river_list(m)
					return // 群组列表
				}
				if len(arg) == 2 && arg[1] == TOOL {
					m.Option(ice.MSG_RIVER, arg[0])
					m.Cmdy(m.Prefix(arg[1]), arg[2:])
					return // 应用列表
				}
				if m.Warn(!m.Right(RIVER, arg), ice.ErrNotRight) {
					return // 没有授权
				}

				switch kit.Select("", arg, 1) {
				case USER, TOOL, NODE:
					m.Option(ice.MSG_RIVER, arg[0])
					m.Cmdy(m.Prefix(arg[1]), arg[2:])
				case kit.MDB_ACTION, aaa.INVITE:
					m.Option(ice.MSG_RIVER, arg[0])
					m.Cmdy(RIVER, arg[1:])
				default:
					m.Cmdy(RIVER, arg)
				}
			}},
		},
	})
}
