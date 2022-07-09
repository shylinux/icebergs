package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const NODE = "node"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		NODE: {Name: "node pod ctx cmd auto insert invite", Help: "设备", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Watch(web.DREAM_START, m.PrefixKey())
				m.Watch(web.SPACE_START, m.PrefixKey())
			}},
			web.SPACE_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
					return
				}
				if msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(mdb.TYPE) == RIVER {
					m.Cmdy(NODE, mdb.INSERT, arg)
				} else {
					msg.Debug(msg.FormatMeta())
				}
			}},
			web.DREAM_START: {Name: "start type name share river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_RIVER, m.Option(RIVER)) == "" {
					return
				}
				m.Cmdy(NODE, mdb.INSERT, arg)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPACE)
			}},
			mdb.INSERT: {Name: "insert type name share river", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, NODE), mdb.HASH, arg)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m, NODE), mdb.HASH, mdb.NAME, m.Option(ice.POD))
			}},
			aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SHARE, m.Cmdx(web.SHARE, mdb.CREATE, mdb.TYPE, RIVER))
				m.Cmdy(code.PUBLISH, ice.CONTEXTS, "tool")
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.OptionFields("time,type,name,share")
				m.Cmdy(mdb.SELECT, RIVER, _river_key(m, NODE), mdb.HASH)
				m.Table(func(index int, value ice.Maps, head []string) {
					m.PushAnchor(value[mdb.NAME], m.MergeURL2("/chat/pod/"+kit.Keys(m.Option(ice.POD), value[mdb.NAME])))
				})
				m.RenameAppend("name", "pod")
				m.PushAction(mdb.REMOVE)
				return
			}
			m.Cmdy(web.ROUTE, arg)
		}},
	}})
}
