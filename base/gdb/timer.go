package gdb

import (
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _timer_create(m *ice.Message, arg ...string) {
	m.Cmdy(mdb.INSERT, TIMER, "", mdb.HASH, DELAY, "10ms", INTERVAL, "10m", ORDER, 1, NEXT, m.Time(m.Option(DELAY)), arg)
}
func _timer_action(m *ice.Message, arg ...string) {
	now := time.Now().UnixNano()
	m.Option(mdb.FIELDS, "time,hash,delay,interval,order,next,cmd")

	m.Richs(TIMER, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		if value = kit.GetMeta(value); value[kit.MDB_STATUS] == cli.STOP {
			return
		}

		order := kit.Int(value[ORDER])
		if n := kit.Time(kit.Format(value[NEXT])); now > n && order > 0 {
			m.Logs(TIMER, kit.MDB_KEY, key, ORDER, order)

			msg := m.Cmd(value[cli.CMD])
			m.Grow(TIMER, kit.Keys(kit.MDB_HASH, key), kit.Dict(cli.RES, msg.Result()))
			if value[ORDER] = kit.Format(order - 1); order > 1 {
				value[NEXT] = msg.Time(value[INTERVAL])
			}
		}
	})
}

const (
	DELAY    = "delay"
	INTERVAL = "interval"
	ORDER    = "order"
	NEXT     = "next"
	TICK     = "tick"
)
const TIMER = "timer"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TIMER: {Name: TIMER, Help: "定时器", Value: kit.Data(TICK, "10ms")},
		},
		Commands: map[string]*ice.Command{
			TIMER: {Name: "timer hash id auto create prunes", Help: "定时器", Action: map[string]*ice.Action{
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
					m.Cmdy(mdb.PRUNES, TIMER, "", mdb.HASH, ORDER, 0)
				}},

				ACTION: {Name: "action", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					_timer_action(m, arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Fields(len(arg), "time,hash,delay,interval,order,next,cmd")
					m.Cmdy(mdb.SELECT, TIMER, "", mdb.HASH, kit.MDB_HASH, arg)
					m.PushAction(mdb.REMOVE)
					return
				}

				m.Fields(len(arg[1:]), "time,id,res")
				m.Cmdy(mdb.SELECT, TIMER, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
			}},
		},
	})
}
