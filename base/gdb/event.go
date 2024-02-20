package gdb

import (
	"sync"

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
			LISTEN: {Name: "listen event* cmd*", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, m.OptionSimple(EVENT, ice.CMD))
				list[m.Option(EVENT)]++
			}},
			HAPPEN: {Name: "happen event*", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				defer m.Cost()
				m.OptionCB(mdb.SELECT, "")
				mdb.ZoneSelect(m.Spawn(ice.OptionFields("")), arg[1]).Table(func(value ice.Maps) {
					m.Cmdy(kit.Split(value[ice.CMD]), arg[1], arg[2:], ice.OptionFields(""))
				})
				_waitMap.Range(func(key, cb ice.Any) bool { cb.(func(*ice.Message, ...string))(m, arg...); return true })
			}},
		}, mdb.ZoneAction(mdb.SHORT, EVENT, mdb.FIELDS, "time,id,cmd"), mdb.ClearOnExitHashAction())},
	})
}

func EventsAction(arg ...string) ice.Actions {
	list := kit.DictList(arg...)
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		for sub := range m.Target().Commands[m.CommandKey()].Actions {
			kit.If(list[sub] == ice.TRUE, func() { Watch(m, sub) })
		}
	}}}
}

var list map[string]int = map[string]int{}

func Watch(m *ice.Message, key string, arg ...string) *ice.Message {
	kit.If(len(arg) == 0, func() { arg = append(arg, m.ShortKey()) })
	return m.Cmd(Prefix(EVENT), LISTEN, EVENT, key, ice.CMD, kit.Join(arg, ice.SP))
}
func Event(m *ice.Message, key string, arg ...ice.Any) *ice.Message {
	if key = kit.Select(kit.Keys(m.CommandKey(), m.ActionKey()), key); list[key] == 0 {
		return m
	}
	return m.Cmdy(Prefix(EVENT), HAPPEN, EVENT, key, arg, logs.FileLineMeta(-1))
}
func EventDeferEvent(m *ice.Message, key string, arg ...ice.Any) func(string, ...ice.Any) {
	Event(m, key, arg...)
	return func(key string, args ...ice.Any) { Event(m, key, args...) }
}

var _waitMap = sync.Map{}

func WaitEvent(m *ice.Message, key string, cb func(*ice.Message, ...string) bool) {
	wg := sync.WaitGroup{}
	h := kit.HashsUniq()
	defer _waitMap.Delete(h)
	_waitMap.Store(h, func(m *ice.Message, arg ...string) {
		m.Info("WaitEvent %v %v", key, kit.FileLine(cb, 3))
		kit.If((key == "" || m.Option(EVENT) == key) && cb(m, arg...), func() { wg.Done() })
	})
	wg.Add(1)
	defer wg.Wait()
}
