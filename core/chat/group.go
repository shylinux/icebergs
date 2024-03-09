package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const GROUP = "group"

func init() {
	Index.MergeCommands(ice.Commands{
		GROUP: {Help: "群组", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.NAME:
					m.AdminCmd(web.MATRIX).Table(func(value ice.Maps) {
						m.Push(arg[0], kit.Keys(kit.Select("", ice.OPS, ice.Info.NodeType == web.WORKER), value[web.DOMAIN], value[mdb.NAME]))
						m.Push(mdb.TYPE, value[mdb.TYPE])
					})
				}
			}},
			mdb.CREATE: {Name: "create type*=worker,server,origin, name*"},
			tcp.SEND: {Name: "send text=hi", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPACE, m.Option(mdb.NAME), Prefix(MESSAGE), tcp.RECV, mdb.TEXT, m.Option(mdb.TEXT))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type*,name*")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(tcp.SEND, mdb.REMOVE)
		}},
	})
}
