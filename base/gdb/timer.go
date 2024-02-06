package gdb

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _timer_action(m *ice.Message, now time.Time, arg ...string) {
	mdb.HashSelects(m).Table(func(value ice.Maps) {
		count := kit.Int(value[mdb.COUNT])
		if count == 0 || value[mdb.TIME] > now.Format(ice.MOD_TIME) {
			return
		}
		m.Options(ice.LOG_DISABLE, ice.FALSE)
		m.Cmd(kit.Split(value[ice.CMD])).Cost()
		kit.If(count < 0, func() { count++ })
		mdb.HashModify(m, mdb.NAME, value[mdb.NAME], mdb.COUNT, count-1, mdb.TIME, m.Time(value[INTERVAL]))
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
		TIMER: {Help: "定时器", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(DELAY, "延时", INTERVAL, "间隔", TICK, "周期")),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.COUNT:
					m.Push(arg[0], "-1")
				case ice.CMD:
					m.Push(arg[0], "cli.procstat insert")
				}
			}},
			mdb.CREATE: {Name: "create name*=hi delay=10ms interval=10s count=3 cmd*=runtime"},
			mdb.PRUNES: {Hand: func(m *ice.Message, arg ...string) { mdb.HashPrunesValue(m, mdb.COUNT, "0") }},
			HAPPEN:     {Hand: func(m *ice.Message, arg ...string) { _timer_action(m, time.Now(), arg...) }},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,delay,interval,count,cmd", TICK, "10s")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).StatusTimeCount(mdb.ConfigSimple(m, TICK))
		}},
	})
}
