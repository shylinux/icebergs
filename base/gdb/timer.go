package gdb

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _timer_action(m *ice.Message, now time.Time, arg ...string) {
	mdb.HashSelects(m).Table(func(value ice.Maps) {
		if value[mdb.COUNT] == "0" {
			return
		}
		if value[mdb.TIME] > now.Format(ice.MOD_TIME) {
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
			mdb.CREATE: {Name: "create name*=hi delay=10ms interval=10s count=3 cmd*=runtime"},
			mdb.PRUNES: {Hand: func(m *ice.Message, arg ...string) { mdb.HashPrunesValue(m, mdb.COUNT, "0") }},
			HAPPEN:     {Hand: func(m *ice.Message, arg ...string) { _timer_action(m, time.Now(), arg...) }},
			RESTART: {Name: "restart count=3", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HashShort(m)), arg)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,name,delay,interval,count,cmd", TICK, "60s"))},
	})
}
