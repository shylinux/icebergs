package gdb

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _timer_action(m *ice.Message, now time.Time, arg ...string) {
	mdb.HashSelects(m).Tables(func(value ice.Maps) {
		if value[mdb.COUNT] == "0" {
			return
		}
		if kit.Time(value[mdb.TIME]) > kit.Int64(now) {
			return
		}
		m.Cmd(ROUTINE, mdb.CREATE, mdb.NAME, value[mdb.NAME], kit.Keycb(ROUTINE), value[ice.CMD])
		mdb.HashModify(m, mdb.HASH, value[mdb.HASH], mdb.COUNT, kit.Int(value[mdb.COUNT])-1, mdb.TIME, m.Time(value[INTERVAL]))
	})
}

const (
	DELAY    = "delay"
	INTERVAL = "interval"
	TICK     = "tick"
)
const TIMER = "timer"

func init() {
	Index.MergeCommands(ice.Commands{
		TIMER: {Name: "timer hash auto create prunes", Help: "定时器", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name=hi delay=10ms interval=10s count=3 cmd=runtime", Help: "创建"},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunesValue(m, mdb.COUNT, "0")
			}},
			HAPPEN: {Name: "happen", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				_timer_action(m, time.Now(), arg...)
			}},
			RESTART: {Name: "restart count=3", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HashShort(m)), arg)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,name,delay,interval,count,cmd", TICK, "10s"))},
	})
}
