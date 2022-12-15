package wx

import (
	ice "shylinux.com/x/icebergs"
)

const EVENT = "event"

func init() {
	Index.MergeCommands(ice.Commands{
		EVENT: {Name: "event", Help: "事件", Actions: ice.Actions{
			"subscribe":   {Help: "订阅", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(MENU, "home") }},
			"unsubscribe": {Help: "取关", Hand: func(m *ice.Message, arg ...string) {}},
		}},
	})
}
