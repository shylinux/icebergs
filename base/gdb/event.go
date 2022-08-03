package gdb

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const EVENT = "event"

func init() {
	Index.MergeCommands(ice.Commands{
		EVENT: {Name: "event event id auto listen happen", Help: "事件流", Actions: ice.MergeAction(ice.Actions{
			LISTEN: {Name: "listen event cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, m.OptionSimple(EVENT, ice.CMD))
			}},
			HAPPEN: {Name: "happen event arg", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneSelect(m, m.Option(EVENT)).Tables(func(value ice.Maps) {
					m.Cmd(kit.Split(value[ice.CMD]), m.Option(EVENT), arg[2:]).Cost()
				})
			}},
		}, mdb.ZoneAction(mdb.SHORT, EVENT, mdb.FIELD, "time,id,cmd"))},
	})
}
func Watch(m *ice.Message, key string, arg ...string) *ice.Message {
	if len(arg) == 0 {
		arg = append(arg, m.Prefix(ice.AUTO))
	}
	return m.Cmd(EVENT, LISTEN, EVENT, key, ice.CMD, kit.Join(arg, ice.SP))
}
func Event(m *ice.Message, key string, arg ...string) *ice.Message {
	return m.Cmd(EVENT, HAPPEN, EVENT, key, arg)
}
