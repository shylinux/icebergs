package gdb

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _event_listen(m *ice.Message, event string, cmd string) {
	h := m.Cmdx(mdb.INSERT, EVENT, "", mdb.HASH, EVENT, event)
	m.Cmdy(mdb.INSERT, EVENT, kit.Keys(kit.MDB_HASH, h), mdb.LIST, ice.CMD, cmd)
}
func _event_action(m *ice.Message, event string, arg ...string) {
	m.Option(mdb.FIELDS, "time,id,cmd")
	m.Cmd(mdb.SELECT, EVENT, kit.Keys(kit.MDB_HASH, kit.Hashs(event)), mdb.LIST).Table(func(index int, value map[string]string, head []string) {
		m.Cmd(kit.Split(value[ice.CMD]), event, arg).Cost(EVENT, event, ice.ARG, arg)
	})
}

const EVENT = "event"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			EVENT: {Name: EVENT, Help: "事件流", Value: kit.Data(kit.MDB_SHORT, EVENT)},
		},
		Commands: map[string]*ice.Command{
			EVENT: {Name: "event event id auto listen", Help: "事件流", Action: map[string]*ice.Action{
				LISTEN: {Name: "listen event cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
					_event_listen(m, m.Option(EVENT), m.Option(ice.CMD))
				}},
				ACTION: {Name: "action event arg", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
					_event_action(m, m.Option(EVENT), arg[2:]...)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, EVENT, "", mdb.HASH, EVENT, m.Option(EVENT))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 { // 事件列表
					m.Fields(len(arg), "time,event,count")
					m.Cmdy(mdb.SELECT, EVENT, "", mdb.HASH)
					m.PushAction(ACTION, mdb.REMOVE)
					return
				}

				m.Fields(len(arg[1:]), "time,id,cmd")
				m.Cmdy(mdb.SELECT, EVENT, kit.Keys(kit.MDB_HASH, kit.Hashs(arg[0])), mdb.LIST, kit.MDB_ID, arg[1:])
			}},
		},
	})
}
