package gdb

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _timer_action(m *ice.Message, arg ...string) {
	now := time.Now().UnixNano()
	m.OptionFields(m.Config(mdb.FIELD))

	m.Richs(TIMER, "", mdb.FOREACH, func(key string, value ice.Map) {
		if value = kit.GetMeta(value); value[cli.STATUS] == cli.STOP {
			return
		}

		order := kit.Int(value[ORDER])
		if n := kit.Time(kit.Format(value[NEXT])); now > n && order > 0 {
			m.Logs(TIMER, mdb.KEY, key, ORDER, order)

			msg := m.Cmd(value[ice.CMD])
			m.Grow(TIMER, kit.Keys(mdb.HASH, key), kit.Dict(ice.RES, msg.Result()))
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
	Index.Merge(&ice.Context{Configs: ice.Configs{
		TIMER: {Name: TIMER, Help: "定时器", Value: kit.Data(
			mdb.FIELD, "time,hash,delay,interval,order,next,cmd", TICK, "1s",
		)},
	}, Commands: ice.Commands{
		TIMER: {Name: "timer hash id auto create action prunes", Help: "定时器", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create delay=10ms interval=10s order=3 cmd=runtime", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, TIMER, "", mdb.HASH, DELAY, "10ms", INTERVAL, "10m", ORDER, 1, NEXT, m.Time(m.Option(DELAY)), arg)
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(mdb.FIELD))
				m.Cmdy(mdb.PRUNES, TIMER, "", mdb.HASH, ORDER, 0)
			}},
			HAPPEN: {Name: "happen", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				_timer_action(m, arg...)
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Fields(len(arg), m.Config(mdb.FIELD), "time,id,res")
			mdb.ZoneSelect(m, arg...)
		}},
	}})
}
