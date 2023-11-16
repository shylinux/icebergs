package team

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
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
	if space := m.Option(web.SPACE); space != "" {
		m.Options(web.SPACE, "").Cmdy(web.SPACE, space, TASK, mdb.MODIFY, field, value, arg)
		return
	}
	if field == STATUS {
		switch value {
		case PROCESS:
			arg = append(arg, BEGIN_TIME, m.Time())
		case CANCEL, FINISH:
			arg = append(arg, END_TIME, m.Time())
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
	END_TIME   = "end_time"

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
		TASK: {Help: "任务", Meta: kit.Dict(
			ctx.TRANS, kit.Dict(html.INPUT, kit.Dict(
				BEGIN_TIME, "起始时间", END_TIME, "结束时间",
				LEVEL, "优先级", SCORE, "完成度",
				PREPARE, "准备中", PROCESS, "进行中", CANCEL, "已取消", FINISH, "已完成",
				ONCE, "一次性", STEP, "阶段性", WEEK, "周期性",
			)),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.ZoneInputs(m, arg); strings.TrimPrefix(arg[0], "extra.") {
				case mdb.STATUS:
					m.Push(arg[0], PREPARE, PROCESS, CANCEL, FINISH)
				case LEVEL, SCORE:
					m.Push(arg[0], "1", "2", "3", "4", "5")
				case mdb.TYPE:
					m.Push(arg[0], ONCE, STEP, WEEK)
				}
				kit.If(arg[0] == mdb.ZONE, func() { m.Push(arg[0], kit.Split(nfs.TemplateText(m, mdb.ZONE))) })
			}},
			mdb.INSERT: {Name: "insert space zone* type*=once,step,week name* text begin_time@date end_time@date", Hand: func(m *ice.Message, arg ...string) {
				if space, arg := arg[1], arg[2:]; space != "" {
					m.Cmdy(web.SPACE, space, TASK, mdb.INSERT, web.SPACE, "", arg)
				} else {
					mdb.ZoneInsert(m, arg[:2], STATUS, PREPARE, LEVEL, 3, SCORE, 3, kit.ArgDef(arg[2:], BEGIN_TIME, m.Time()))
				}
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, arg[0], arg[1], arg[2:]...) }},
			CANCEL:     {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, CANCEL) }},
			BEGIN:      {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, PROCESS) }},
			END:        {Hand: func(m *ice.Message, arg ...string) { _task_modify(m, STATUS, FINISH) }},
			web.STATS_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn()); msg.Length() > 0 {
					count := 0
					msg.Table(func(value ice.Maps) { count += kit.Int(value[mdb.COUNT]) })
					web.PushStats(m, kit.Keys(m.CommandKey(), mdb.TOTAL), count, "")
				}
			}},
		}, web.StatsAction(), mdb.ExportZoneAction(mdb.FIELD, "time,zone,count", mdb.FIELDS, "begin_time,end_time,id,status,level,score,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) > 0 && arg[0] != "" {
				status := map[string]int{}
				m.Table(func(value ice.Maps) { m.PushButton(_task_action(m, value[STATUS])) })
				m.Table(func(value ice.Maps) { status[value[mdb.STATUS]]++ }).StatusTimeCount(status)
			}
		}},
	})
}
