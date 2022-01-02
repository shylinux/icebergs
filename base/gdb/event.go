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
	mdb.ZoneSelect(m, event).Table(func(index int, value map[string]string, head []string) {
		m.Cmd(kit.Split(value[ice.CMD]), event, arg).Cost(EVENT, event, ice.ARG, arg)
	})
}

const EVENT = "event"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		EVENT: {Name: EVENT, Help: "事件流", Value: kit.Data(
			mdb.SHORT, EVENT, mdb.FIELD, "time,id,cmd",
		)},
	}, Commands: map[string]*ice.Command{
		EVENT: {Name: "event event id auto listen", Help: "事件流", Action: ice.MergeAction(map[string]*ice.Action{
			LISTEN: {Name: "listen event cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				_event_listen(m, m.Option(EVENT), m.Option(ice.CMD))
			}},
			ACTION: {Name: "action event arg", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				_event_action(m, m.Option(EVENT), arg[2:]...)
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(ACTION, mdb.REMOVE)
			}
		}},
	}})
}
