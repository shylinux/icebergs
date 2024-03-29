package lark

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	P2P_CHAT_CREATE = "p2p_chat_create"
	MESSAGE_READ    = "message_read"
	CHAT_DISBAND    = "chat_disband"
	ADD_BOT         = "add_bot"
	MSG_TYPE        = "msg_type"
)
const EVENT = "event"

func init() {
	Index.MergeCommands(ice.Commands{
		EVENT: {Name: "event", Help: "事件", Actions: ice.Actions{
			P2P_CHAT_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(OPEN_CHAT_ID) != "" {
					m.Cmdy(SEND, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), mdb.Conf(m, APP, kit.Keym(nfs.TEMPLATE, m.Option(mdb.TYPE))))
				}
			}},
			MESSAGE_READ: {Hand: func(m *ice.Message, arg ...string) {}},
			CHAT_DISBAND: {Hand: func(m *ice.Message, arg ...string) {}},
			ADD_BOT: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(OPEN_CHAT_ID) != "" {
					m.Cmdy(SEND, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), mdb.Conf(m, APP, kit.Keym(nfs.TEMPLATE, m.Option(mdb.TYPE))))
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(MSG, m.Option(MSG_TYPE)) }},
	})
}
