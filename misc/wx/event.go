package wx

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const EVENT = "event"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		EVENT: {Name: EVENT, Help: "事件", Value: kit.Data()},
	}, Commands: ice.Commands{
		EVENT: {Name: "event", Help: "事件", Actions: ice.Actions{
			"subscribe": {Name: "subscribe", Help: "订阅", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(MENU, "home")
			}},
			"unsubscribe": {Name: "unsubscribe", Help: "取关", Hand: func(m *ice.Message, arg ...string) {
			}},
		}},
	}})
}
