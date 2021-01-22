package team

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"strings"
	"time"
)

func _sub_key(m *ice.Message, zone string) string {
	return kit.Keys(kit.MDB_HASH, kit.Hashs(zone))
}
func _task_scope(m *ice.Message, tz int, arg ...string) (time.Time, time.Time) {
	begin_time := time.Now()
	if len(arg) > 1 {
		begin_time, _ = time.ParseInLocation(ice.MOD_TIME, arg[1], time.Local)
	}

	begin_time = begin_time.Add(time.Duration(tz) * time.Hour)
	begin_time = begin_time.Add(-time.Duration(begin_time.UnixNano()) % (24 * time.Hour))
	begin_time = begin_time.Add(-time.Duration(tz) * time.Hour)

	end_time := begin_time
	switch arg[0] {
	case TaskScale.DAY:
		end_time = begin_time.AddDate(0, 0, 1)
	case TaskScale.WEEK:
		begin_time = begin_time.AddDate(0, 0, -int(begin_time.Weekday()))
		end_time = begin_time.AddDate(0, 0, 7)
	case TaskScale.MONTH:
		begin_time = begin_time.AddDate(0, 0, -begin_time.Day()+1)
		end_time = begin_time.AddDate(0, 1, 0)
	case TaskScale.YEAR:
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1)
		end_time = begin_time.AddDate(1, 0, 0)
	case TaskScale.LONG:
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1)
		begin_time = begin_time.AddDate(-5, 0, 0)
		end_time = begin_time.AddDate(10, 0, 0)
	}

	return begin_time, end_time
}
func _task_action(m *ice.Message, status interface{}, action ...string) string {
	switch status {
	case TaskStatus.PREPARE:
		action = append(action, gdb.BEGIN)
	case TaskStatus.PROCESS:
		action = append(action, gdb.END)
	case TaskStatus.CANCEL:
	case TaskStatus.FINISH:
	}
	return strings.Join(action, ",")
}

func _task_list(m *ice.Message, zone string, id string) {
	if zone == "" {
		m.Option(mdb.FIELDS, "time,zone,count")
	} else {
		m.Option(mdb.FIELDS, kit.Select("begin_time,id,status,level,score,type,name,text", mdb.DETAIL, id != ""))
		defer m.Table(func(index int, value map[string]string, head []string) {
			m.PushButton(_task_action(m, value[TaskField.STATUS]))
		})
	}
	m.Cmdy(mdb.SELECT, TASK, "", mdb.ZONE, zone, id)
}
func _task_create(m *ice.Message, zone string) {
	m.Cmdy(mdb.INSERT, TASK, "", mdb.HASH, kit.MDB_ZONE, zone)
}
func _task_insert(m *ice.Message, zone string, arg ...string) {
	m.Cmdy(mdb.INSERT, TASK, _sub_key(m, zone), mdb.LIST,
		TaskField.BEGIN_TIME, m.Time(), TaskField.CLOSE_TIME, m.Time("30m"),
		TaskField.STATUS, TaskStatus.PREPARE, TaskField.LEVEL, 3, TaskField.SCORE, 3,
		arg,
	)
}
func _task_modify(m *ice.Message, zone, id, field, value string, arg ...string) {
	if field == TaskField.STATUS {
		switch value {
		case TaskStatus.PROCESS:
			arg = append(arg, TaskField.BEGIN_TIME, m.Time())
		case TaskStatus.CANCEL, TaskStatus.FINISH:
			arg = append(arg, TaskField.CLOSE_TIME, m.Time())
		}
	}
	m.Cmdy(mdb.MODIFY, TASK, _sub_key(m, zone), mdb.LIST, kit.MDB_ID, id, field, value, arg)
}
func _task_delete(m *ice.Message, zone, id string) {
	m.Cmdy(mdb.MODIFY, TASK, _sub_key(m, zone), mdb.LIST, kit.MDB_ID, id, TaskField.STATUS, TaskStatus.CANCEL)
}
func _task_export(m *ice.Message, file string) {
	m.Option(mdb.FIELDS, "zone,id,time,type,name,text,level,status,score,begin_time,close_time,extra")
	m.Cmdy(mdb.EXPORT, TASK, "", mdb.ZONE, file)
}
func _task_import(m *ice.Message, file string) {
	m.Option(mdb.FIELDS, "zone")
	m.Cmdy(mdb.IMPORT, TASK, "", mdb.ZONE, file)
}
func _task_inputs(m *ice.Message, field, value string) {
	switch field {
	case kit.MDB_ZONE:
		m.Cmdy(mdb.INPUTS, TASK, "", mdb.HASH, field, value)
	default:
		m.Cmdy(mdb.INPUTS, TASK, _sub_key(m, m.Option(kit.MDB_ZONE)), mdb.LIST, field, value)
	}
}
func _task_search(m *ice.Message, kind, name, text string) {
	if kind != TASK && kind != kit.MDB_FOREACH {
		return
	}

	m.Cmd(mdb.SELECT, m.Prefix(TASK), "", mdb.ZONE, kit.MDB_FOREACH, func(key string, value map[string]interface{}, val map[string]interface{}) {
		if name != "" && !kit.Contains(value[kit.MDB_NAME], name) {
			return
		}
		if kind == TASK {
			m.PushSearch(kit.SSH_CMD, TASK,
				kit.MDB_ZONE, val[kit.MDB_ZONE], kit.MDB_ID, kit.Format(value[kit.MDB_ID]),
				value)
		} else {
			m.PushSearch(kit.SSH_CMD, TASK,
				kit.MDB_TYPE, val[kit.MDB_ZONE], kit.MDB_NAME, kit.Format(value[kit.MDB_ID]),
				kit.MDB_TEXT, kit.Format("%v:%v", value[kit.MDB_NAME], value[kit.MDB_TEXT]),
				value)
		}
	})
}

var TaskField = struct{ LEVEL, STATUS, SCORE, BEGIN_TIME, CLOSE_TIME string }{
	LEVEL:  "level",
	STATUS: "status",
	SCORE:  "score",

	BEGIN_TIME: "begin_time",
	CLOSE_TIME: "close_time",
}
var TaskStatus = struct{ PREPARE, PROCESS, CANCEL, FINISH string }{
	PREPARE: "prepare",
	PROCESS: "process",
	CANCEL:  "cancel",
	FINISH:  "finish",
}
var TaskType = struct{ ONCE, STEP, WEEK string }{
	ONCE: "once",
	STEP: "step",
	WEEK: "week",
}
var TaskScale = struct{ DAY, WEEK, MONTH, YEAR, LONG string }{
	DAY:   "day",
	WEEK:  "week",
	MONTH: "month",
	YEAR:  "year",
	LONG:  "long",
}

const TASK = "task"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TASK: {Name: TASK, Help: "任务", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
		},
		Commands: map[string]*ice.Command{
			TASK: {Name: "task zone id auto insert export import", Help: "任务", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_task_create(m, arg[1])
					_task_insert(m, arg[1], arg[2:]...)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1])
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					_task_delete(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID))
				}},
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					_task_export(m, m.Option(kit.MDB_FILE))
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_task_import(m, m.Option(kit.MDB_FILE))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					_task_inputs(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				}},
				mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_task_search(m, arg[0], arg[1], arg[2])
					m.PushPodCmd(TASK, kit.Simple(mdb.SEARCH, arg)...)
				}},

				gdb.BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), TaskField.STATUS, TaskStatus.PROCESS)
				}},
				gdb.END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), TaskField.STATUS, TaskStatus.FINISH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(kit.SSH_POD) != "" {
					m.Cmdy(web.SPACE, m.Option(kit.SSH_POD), m.Prefix(TASK), arg)
					return
				}
				_task_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
			}},
		},
	})
}
