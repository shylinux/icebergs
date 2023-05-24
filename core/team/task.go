package team

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _task_action(m *ice.Message, status ice.Any, action ...string) string {
	switch status {
	case PREPARE:
		action = append(action, BEGIN, CANCEL)
	case PROCESS:
		action = append(action, END, CANCEL)
	case CANCEL:
		action = append(action, BEGIN)
	case FINISH:
	}
	return kit.Join(action)
}
func _task_modify(m *ice.Message, field, value string, arg ...string) {
	if field == STATUS {
		switch value {
		case PROCESS:
			arg = append(arg, BEGIN_TIME, m.Time())
		case CANCEL, FINISH:
			arg = append(arg, CLOSE_TIME, m.Time())
		}
	}
	mdb.ZoneModify(m, field, value, arg)
}

const (
	ONCE = "once"
	STEP = "step"
	// WEEK = "week"
)
const (
	PREPARE = "prepare"
	PROCESS = "process"
	CANCEL  = "cancel"
	FINISH  = "finish"
)
const (
	BEGIN_TIME = "begin_time"
	CLOSE_TIME = "close_time"

	STATUS = "status"
	LEVEL  = "level"
	SCORE  = "score"
)
const (
	BEGIN = "begin"
	END   = "end"
)
const TASK = "task"

func init() {
	Index.MergeCommands(ice.Commands{
		TASK: {Name: "task zone id auto insert", Help: "任务", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] = strings.TrimPrefix(arg[0], "extra."); arg[0] {
				case mdb.STATUS:
					m.Push(arg[0], PREPARE, PROCESS, CANCEL, FINISH)
				case LEVEL, SCORE:
					m.Push(arg[0], "1", "2", "3", "4", "5")
				case mdb.TYPE:
					m.Push(arg[0], ONCE, STEP, WEEK)
				case ctx.INDEX, ctx.ARGS:
					m.Option(ctx.INDEX, m.Option("extra.index"))
					ctx.CmdInputs(m, arg...)
				default:
					mdb.ZoneInputs(m, arg)
				}
			}},
			mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, arg[:2], BEGIN_TIME, m.Time(), STATUS, PREPARE, LEVEL, 3, SCORE, 3, arg[2:])
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, arg[0], arg[1], arg[2:]...) }},
			CANCEL:     {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, CANCEL) }},
			BEGIN:      {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, PROCESS) }},
			END:        {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, FINISH) }},
		}, mdb.ImportantZoneAction(mdb.FIELD, "begin_time,close_time,id,status,level,score,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) > 0 && arg[0] != "" {
				status := map[string]int{}
				m.Table(func(value ice.Maps) { m.PushButton(_task_action(m, value[STATUS])) })
				m.Table(func(value ice.Maps) { status[value[mdb.STATUS]]++ }).StatusTimeCount(status)
			}
		}},
	})
}
