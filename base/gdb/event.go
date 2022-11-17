package gdb

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const EVENT = "event"

func init() {
	Index.MergeCommands(ice.Commands{
		EVENT: {Name: "event event id auto listen happen", Help: "事件流", Actions: ice.MergeActions(ice.Actions{
			LISTEN: {Name: "listen event cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, m.OptionSimple(EVENT, ice.CMD))
			}},
			HAPPEN: {Name: "happen event arg", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneSelect(m.Spawn(ice.OptionFields("")), m.Option(EVENT)).Tables(func(value ice.Maps) {
					m.Cmdy(kit.Split(value[ice.CMD]), m.Option(EVENT), arg[2:], ice.OptionFields("")).Cost()
				})
			}},
		}, mdb.ZoneAction(mdb.SHORT, EVENT, mdb.FIELD, "time,id,cmd"))},
	})
}
func EventAction(arg ...string) ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, args ...string) {
			for _, v := range arg {
				Watch(m, v)
			}
		}},
	}
}
func Watch(m *ice.Message, key string, arg ...string) *ice.Message {
	if len(arg) == 0 {
		arg = append(arg, m.PrefixKey())
	}
	return m.Cmd("gdb.event", LISTEN, EVENT, key, ice.CMD, kit.Join(arg, ice.SP))
}
func Event(m *ice.Message, key string, arg ...ice.Any) *ice.Message {
	return m.Cmdy("gdb.event", HAPPEN, EVENT, kit.Select(kit.Keys(m.CommandKey(), m.ActionKey()), key), arg)
}
