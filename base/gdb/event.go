package gdb

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

const (
	LISTEN = "listen"
	HAPPEN = "happen"
)
const EVENT = "event"

func init() {
	Index.MergeCommands(ice.Commands{
		EVENT: {Name: "event event id auto listen happen", Help: "事件流", Actions: ice.MergeActions(ice.Actions{
			LISTEN: {Name: "listen event cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, m.OptionSimple(EVENT, ice.CMD))
			}},
			HAPPEN: {Name: "happen event", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				defer m.Cost()
				mdb.ZoneSelect(m.Spawn(ice.OptionFields("")), m.Option(EVENT)).Tables(func(value ice.Maps) {
					m.Cmdy(kit.Split(value[ice.CMD]), m.Option(EVENT), arg[2:], ice.OptionFields(""))
				})
			}},
		}, mdb.ZoneAction(mdb.SHORT, EVENT, mdb.FIELD, "time,id,cmd"))},
	})
}
func EventAction(arg ...string) ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, _ ...string) {
			for _, v := range arg {
				Watch(m, v)
			}
		}},
	}
}
func EventsAction(arg ...string) ice.Actions {
	list := kit.DictList(arg...)
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		for sub := range m.Target().Commands[m.CommandKey()].Actions {
			if list[sub] == ice.TRUE {
				Watch(m, sub)
			}
		}
	}}}
}
func Watch(m *ice.Message, key string, arg ...string) *ice.Message {
	if len(arg) == 0 {
		arg = append(arg, m.PrefixKey())
	}
	return m.Cmd(EVENT, LISTEN, EVENT, key, ice.CMD, kit.Join(arg, ice.SP))
}
func Event(m *ice.Message, key string, arg ...ice.Any) *ice.Message {
	return m.Cmdy(EVENT, HAPPEN, EVENT, kit.Select(kit.Keys(m.CommandKey(), m.ActionKey()), key), arg, logs.FileLineMeta(logs.FileLine(-1)))
}
