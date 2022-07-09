package gdb

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _event_listen(m *ice.Message, event string, cmd string) {
	m.Cmdy(mdb.INSERT, EVENT, "", mdb.HASH, EVENT, event)
	m.Cmdy(mdb.INSERT, EVENT, "", mdb.ZONE, event, ice.CMD, cmd)
}
func _event_action(m *ice.Message, event string, arg ...string) {
	mdb.ZoneSelect(m, event).Table(func(index int, value ice.Maps, head []string) {
		m.Cmd(kit.Split(value[ice.CMD]), event, arg).Cost(EVENT, event, ice.ARG, arg)
	})
}

const EVENT = "event"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		EVENT: {Name: EVENT, Help: "事件流", Value: kit.Data(mdb.SHORT, EVENT, mdb.FIELD, "time,id,cmd")},
	}, Commands: ice.Commands{
		EVENT: {Name: "event event id auto listen", Help: "事件流", Actions: ice.MergeAction(ice.Actions{
			LISTEN: {Name: "listen event cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				_event_listen(m, m.Option(EVENT), m.Option(ice.CMD))
			}},
			HAPPEN: {Name: "happen event arg", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				_event_action(m, m.Option(EVENT), arg[2:]...)
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(HAPPEN, mdb.REMOVE)
			}
		}},
	}})
}
