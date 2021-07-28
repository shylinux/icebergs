package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const NODE = "node"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Watch(web.SPACE_START, m.Prefix(NODE))
			}},
			NODE: {Name: "node name ctx cmd auto insert invite", Help: "设备", Action: map[string]*ice.Action{
				web.SPACE_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
						return
					}
					if msg := m.Cmd(AUTH, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) == NODE {
						m.Cmdy(mdb.INSERT, RIVER, _river_key(m, NODE), mdb.HASH, arg)
					}
				}},
				aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					m.Option(web.SHARE, m.Cmdx(AUTH, mdb.CREATE, kit.MDB_TYPE, NODE))
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
						m.PushAnchor(value[kit.MDB_NAME], kit.MergeURL(m.Option(ice.MSG_USERWEB),
							cli.POD, kit.Keys(m.Option(cli.POD), value[kit.MDB_NAME])))
					})
					m.PushAction(mdb.REMOVE)
					return
				}
				m.Cmdy(web.ROUTE, arg)
			}},
		},
	})
}
