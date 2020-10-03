package gdb

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _event_listen(m *ice.Message, event string, cmd string) {
	h := m.Cmdx(mdb.INSERT, EVENT, "", mdb.HASH, EVENT, event)
	m.Cmdy(mdb.INSERT, EVENT, kit.Keys(kit.MDB_HASH, h), mdb.LIST, kit.SSH_CMD, cmd)
}
func _event_action(m *ice.Message, event string, arg ...string) {
	m.Option(mdb.FIELDS, "time,id,cmd")
	m.Cmd(mdb.SELECT, EVENT, kit.Keys(kit.MDB_HASH, kit.Hashs(event)), mdb.LIST).Table(func(index int, value map[string]string, head []string) {
		m.Cmd(kit.Split(value[kit.SSH_CMD]), event, arg).Cost("event %v %v", event, arg)
	})
}

const EVENT = "event"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			EVENT: {Name: EVENT, Help: "事件流", Value: kit.Data(
				kit.MDB_SHORT, EVENT,
			)},
		},
		Commands: map[string]*ice.Command{
			EVENT: {Name: "event event id auto 监听", Help: "事件流", Action: map[string]*ice.Action{
				LISTEN: {Name: "listen event cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
					_event_listen(m, m.Option(EVENT), m.Option(kit.SSH_CMD))
				}},
				ACTION: {Name: "action event arg", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
					_event_action(m, m.Option(EVENT), arg[2:]...)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, EVENT, "", mdb.HASH, EVENT, m.Option(EVENT))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,event,count")
					m.Cmdy(mdb.SELECT, EVENT, "", mdb.HASH)
					m.PushAction(ACTION, mdb.REMOVE)
					return
				}

				m.Option(mdb.FIELDS, kit.Select("time,id,cmd", mdb.DETAIL, len(arg) > 1))
				m.Cmdy(mdb.SELECT, EVENT, kit.Keys(kit.MDB_HASH, kit.Hashs(arg[0])), mdb.LIST, kit.MDB_ID, arg[1:])
				if len(arg) == 1 {
					m.Sort(kit.MDB_ID)
				}
			}},
		},
	}, nil)
}
