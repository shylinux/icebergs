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
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		EVENT: {Name: "event", Help: "事件", Action: map[string]*ice.Action{
			P2P_CHAT_CREATE: {Name: "", Help: "", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(OPEN_CHAT_ID) != "" {
					m.Cmdy(SEND, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), m.Conf(APP, kit.Keym(nfs.TEMPLATE, m.Option(mdb.TYPE))))
				}
			}},
			MESSAGE_READ: {Name: "", Help: "", Hand: func(m *ice.Message, arg ...string) {
			}},
			CHAT_DISBAND: {Name: "", Help: "", Hand: func(m *ice.Message, arg ...string) {
			}},
			ADD_BOT: {Name: "", Help: "", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(OPEN_CHAT_ID) != "" {
					m.Cmdy(SEND, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), m.Conf(APP, kit.Keym(nfs.TEMPLATE, m.Option(mdb.TYPE))))
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(MSG, m.Option(MSG_TYPE))
		}},
	}})
}
