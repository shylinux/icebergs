package wx

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const EVENT = "event"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		EVENT: {Name: EVENT, Help: "事件", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		EVENT: {Name: "event", Help: "事件", Action: map[string]*ice.Action{
			"subscribe": {Name: "subscribe", Help: "订阅", Hand: func(m *ice.Message, arg ...string) {
				_wx_action(m.Cmdy(MENU, "home"))
			}},
			"unsubscribe": {Name: "unsubscribe", Help: "取关", Hand: func(m *ice.Message, arg ...string) {
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	}})
}
