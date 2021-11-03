package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _task_action(m *ice.Message, status interface{}, action ...string) string {
	switch status {
	case PREPARE:
		action = append(action, BEGIN)
	case PROCESS:
		action = append(action, END)
	case CANCEL, FINISH:
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
	m.Cmdy(mdb.MODIFY, m.Prefix(TASK), "", mdb.ZONE, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), field, value, arg)
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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TASK: {Name: TASK, Help: "任务", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "begin_time,id,status,level,score,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		TASK: {Name: "task zone id auto insert export import", Help: "任务", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(TASK), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
				m.Cmdy(mdb.INSERT, m.Prefix(TASK), "", mdb.ZONE, m.Option(kit.MDB_ZONE),
					BEGIN_TIME, m.Time(), CLOSE_TIME, m.Time("30m"),
					STATUS, PREPARE, LEVEL, 3, SCORE, 3, arg)
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, arg[0], arg[1])
				m.ProcessRefresh30ms()
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, STATUS, CANCEL)
			}},
			mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(kit.MDB_ZONE, "time,id,type,name,text,level,status,score,begin_time,close_time,extra")
				m.Cmdy(mdb.EXPORT, m.Prefix(TASK), "", mdb.ZONE)
				m.ProcessRefresh30ms()
			}},

			BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, STATUS, PROCESS)
			}},
			END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, STATUS, FINISH)
			}},
		}, mdb.ZoneAction(), ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(mdb.REMOVE)
			} else {
				status := map[string]int{}
				m.Table(func(index int, value map[string]string, head []string) {
					m.PushButton(_task_action(m, value[STATUS]))
					status[value[kit.MDB_STATUS]]++
				})
				m.Status(status)
			}
		}},
	}})
}
