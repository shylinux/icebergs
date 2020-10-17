package gdb

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"time"
)

func _timer_create(m *ice.Message, arg ...string) {
	m.Cmdy(mdb.INSERT, TIMER, "", mdb.HASH, "delay", "10ms", "interval", "10m", "order", 1, "next", m.Time(m.Option("delay")), arg)
}
func _timer_action(m *ice.Message, arg ...string) {
	now := time.Now().UnixNano()
	m.Option(mdb.FIELDS, "time,hash,delay,interval,order,next,cmd")

	m.Richs(TIMER, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		if value = kit.GetMeta(value); value[kit.MDB_STATUS] == STOP {
			return
		}

		order := kit.Int(value["order"])
		if n := kit.Time(kit.Format(value["next"])); now > n && order > 0 {
			m.Logs(TIMER, "key", key, "order", order)

			msg := m.Cmd(value[kit.SSH_CMD])
			m.Grow(TIMER, kit.Keys(kit.MDB_HASH, key), kit.Dict("res", msg.Result()))
			if value["order"] = kit.Format(order - 1); order > 1 {
				value["next"] = msg.Time(value["interval"])
			}
		}
	})
}

const TIMER = "timer"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TIMER: {Name: TIMER, Help: "定时器", Value: kit.Data("tick", "100ms")},
		},
		Commands: map[string]*ice.Command{
			TIMER: {Name: "timer hash id auto 添加 清理", Help: "定时器", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create delay=10ms interval=10s order=3 cmd=runtime", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_timer_create(m, arg...)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, TIMER, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, TIMER, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,delay,interval,order,next,cmd")
					m.Cmdy(mdb.PRUNES, TIMER, "", mdb.HASH, "order", 0)
				}},

				ACTION: {Name: "action", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					_timer_action(m, arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, kit.Select("time,hash,delay,interval,order,next,cmd", mdb.DETAIL, len(arg) > 0))
					m.Cmdy(mdb.SELECT, TIMER, "", mdb.HASH, kit.MDB_HASH, arg)
					m.PushAction(mdb.REMOVE)
					return
				}

				m.Option(mdb.FIELDS, kit.Select("time,id,res", mdb.DETAIL, len(arg) > 1))
				m.Cmdy(mdb.SELECT, TIMER, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				return
			}},
		},
	}, nil)
}