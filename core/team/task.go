package team

import (
	ice "shylinux.com/x/icebergs"
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

const ( // type
	ONCE = "once"
	STEP = "step"
	// WEEK = "week"
)
const ( // status
	PREPARE = "prepare"
	PROCESS = "process"
	CANCEL  = "cancel"
	FINISH  = "finish"
)
const ( // key
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
		TASK: {Name: "task zone id auto insert export import", Help: "任务", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, arg[:2], BEGIN_TIME, m.Time(), STATUS, PREPARE, LEVEL, 3, SCORE, 3, arg[2:])
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, arg[0], arg[1], arg[2:]...) }},
			CANCEL: {Name: "cancal", Help: "取消", Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, CANCEL) }},
			BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, PROCESS) }},
			END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, FINISH) }},
		}, mdb.ZoneAction(mdb.FIELD, "begin_time,close_time,id,status,level,score,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) > 0 && arg[0] != "" {
				status := map[string]int{}
				m.Tables(func(value ice.Maps) {
					m.PushButton(_task_action(m, value[STATUS]))
					status[value[mdb.STATUS]]++
				})
				m.StatusTimeCount(status)
			}
		}},
	})
}
