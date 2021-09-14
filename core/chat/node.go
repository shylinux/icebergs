package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const NODE = "node"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Watch(web.DREAM_START, m.Prefix(NODE))
				m.Watch(web.SPACE_START, m.Prefix(NODE))
			}},
			NODE: {Name: "node pod ctx cmd auto insert invite", Help: "设备", Action: map[string]*ice.Action{
				web.DREAM_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
						return
					}
					m.Cmdy(mdb.INSERT, RIVER, _river_key(m, NODE), mdb.HASH, arg)
				}},
				web.SPACE_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
						return
					}
					if msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) == RIVER {
						m.Cmdy(mdb.INSERT, RIVER, _river_key(m, NODE), mdb.HASH, arg)
					}
				}},
				aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					m.Option(web.SHARE, m.Cmdx(web.SHARE, mdb.CREATE, kit.MDB_TYPE, RIVER))
					m.Cmdy(code.PUBLISH, ice.CONTEXTS, "tool")
				}},
				mdb.INSERT: {Name: "insert type name share", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, RIVER, _river_key(m, NODE), mdb.HASH, arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, RIVER, _river_key(m, NODE), mdb.HASH, m.OptionSimple(aaa.USERNAME))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.SPACE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,type,name,share")
					m.Cmdy(mdb.SELECT, RIVER, kit.Keys(kit.MDB_HASH, m.Option(ice.MSG_RIVER), NODE), mdb.HASH)
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushAnchor(value[kit.MDB_NAME], kit.MergeURL2(m.Option(ice.MSG_USERWEB),
							"/chat/pod/"+kit.Keys(m.Option(cli.POD), value[kit.MDB_NAME])))
					})
					m.RenameAppend("name", "pod")
					m.PushAction(mdb.REMOVE)
					return
				}
				m.Cmdy(web.ROUTE, arg)
			}},
		},
	})
}