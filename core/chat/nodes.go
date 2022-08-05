package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const NODES = "nodes"

func init() {
	Index.MergeCommands(ice.Commands{
		NODES: {Name: "nodes space index auto insert invite", Help: "设备", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Watch(m, web.SPACE_START, m.PrefixKey())
				gdb.Watch(m, web.DREAM_START, m.PrefixKey())
			}},
			web.SPACE_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
					return
				}
				if msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(mdb.TYPE) == RIVER {
					m.Cmdy("", mdb.INSERT, arg)
				}
			}},
			web.DREAM_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
					return
				}
				m.Cmdy("", mdb.INSERT, arg)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPACE)
			}},
			mdb.INSERT: {Name: "insert type space share river", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.SHORT, web.SPACE)
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, m.CommandKey()), mdb.HASH, arg)
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m, m.CommandKey()), mdb.HASH, m.OptionSimple(web.SPACE))
			}},
			aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SHARE, m.Cmdx(web.SHARE, mdb.CREATE, mdb.TYPE, RIVER))
				m.Cmdy("publish", ice.CONTEXTS, "tool")
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.OptionFields("time,type,space,share")
				m.Cmdy(mdb.SELECT, RIVER, _river_key(m, m.CommandKey()), mdb.HASH)
				m.Tables(func(value ice.Maps) {
					m.PushAnchor(value[mdb.NAME], web.MergePod(m, kit.Keys(m.Option(ice.POD), value[mdb.NAME])))
				})
				m.PushAction(mdb.REMOVE)
				return
			}
			m.Cmdy(web.ROUTE, arg)
		}},
	})
}
