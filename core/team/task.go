package team

import (
	"strings"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _task_scope(m *ice.Message, tz int, arg ...string) (time.Time, time.Time) {
	begin_time := time.Now()
	if len(arg) > 1 {
		begin_time, _ = time.ParseInLocation(ice.MOD_TIME, arg[1], time.Local)
	}

	begin_time = begin_time.Add(time.Duration(tz) * time.Hour)
	begin_time = begin_time.Add(-time.Duration(begin_time.UnixNano()) % (24 * time.Hour))
	begin_time = begin_time.Add(-time.Duration(tz) * time.Hour)

	end_time := begin_time
	switch kit.Select(WEEK, arg, 0) {
	case DAY:
		end_time = begin_time.AddDate(0, 0, 1)
	case WEEK:
		begin_time = begin_time.AddDate(0, 0, -int(begin_time.Weekday()))
		end_time = begin_time.AddDate(0, 0, 7)
	case MONTH:
		begin_time = begin_time.AddDate(0, 0, -begin_time.Day()+1)
		end_time = begin_time.AddDate(0, 1, 0)
	case YEAR:
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1)
		end_time = begin_time.AddDate(1, 0, 0)
	case LONG:
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1)
		begin_time = begin_time.AddDate(-5, 0, 0)
		end_time = begin_time.AddDate(10, 0, 0)
	}

	return begin_time, end_time
}
func _task_action(m *ice.Message, status interface{}, action ...string) string {
	switch status {
	case PREPARE:
		action = append(action, BEGIN)
	case PROCESS:
		action = append(action, END)
	case CANCEL:
	case FINISH:
	}
	return strings.Join(action, ",")
}

func _task_create(m *ice.Message, zone string) {
	m.Cmdy(mdb.INSERT, TASK, "", mdb.HASH, kit.MDB_ZONE, zone)
}
func _task_insert(m *ice.Message, zone string, arg ...string) {
	m.Cmdy(mdb.INSERT, TASK, kit.KeyHash(zone), mdb.LIST,
		BEGIN_TIME, m.Time(), CLOSE_TIME, m.Time("30m"),
		STATUS, PREPARE, LEVEL, 3, SCORE, 3, arg)
}
func _task_modify(m *ice.Message, zone, id, field, value string, arg ...string) {
	if field == STATUS {
		switch value {
		case PROCESS:
			arg = append(arg, BEGIN_TIME, m.Time())
		case CANCEL, FINISH:
			arg = append(arg, CLOSE_TIME, m.Time())
		}
	}
	m.Cmdy(mdb.MODIFY, TASK, "", mdb.ZONE, zone, id, field, value, arg)
}
func _task_inputs(m *ice.Message, field, value string) {
	switch strings.TrimPrefix(field, "extra.") {
	case "pod":
		m.Cmd(web.SPACE).Table(func(index int, value map[string]string, head []string) {
			m.Push(field, value[kit.MDB_NAME])
			m.Push("", value, []string{kit.MDB_TYPE})
		})
	case "ctx":
		m.Cmd(m.Space(m.Option("extra.pod")), ctx.CONTEXT).Table(func(index int, value map[string]string, head []string) {
			m.Push(field, value[kit.MDB_NAME])
			m.Push("", value, []string{kit.MDB_HELP})
		})
	case "cmd":
		m.Cmd(m.Space(m.Option("extra.pod")), ctx.CONTEXT, m.Option("extra.ctx"), ctx.COMMAND).Table(func(index int, value map[string]string, head []string) {
			m.Push(field, value[kit.MDB_KEY])
			m.Push("", value, []string{kit.MDB_HELP})
		})
	case "arg":

	case kit.MDB_ZONE:
		m.Cmdy(mdb.INPUTS, TASK, "", mdb.HASH, field, value)
	default:
		m.Cmdy(mdb.INPUTS, TASK, kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.LIST, field, value)
	}
}
func _task_search(m *ice.Message, kind, name, text string) {
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

const ( // type
	ONCE = "once"
	STEP = "step"
)
const ( // scale
	DAY   = "day"
	WEEK  = "week"
	MONTH = "month"
	YEAR  = "year"
	LONG  = "long"
)
const ( // status
	PREPARE = "prepare"
	PROCESS = "process"
	CANCEL  = "cancel"
	FINISH  = "finish"
)
const ( // key
	SCALE  = "scale"
	LEVEL  = "level"
	STATUS = "status"
	SCORE  = "score"

	BEGIN_TIME = "begin_time"
	CLOSE_TIME = "close_time"
)

const TASK = "task"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TASK: {Name: TASK, Help: "任务", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "begin_time,id,status,level,score,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, TASK, m.Prefix(TASK))
			}},
			TASK: {Name: "task zone id auto insert export import", Help: "任务", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_task_create(m, arg[1])
					_task_insert(m, arg[1], arg[2:]...)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1])
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, CANCEL)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, TASK, "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(kit.MDB_ZONE, "time,id,type,name,text,level,status,score,begin_time,close_time,extra")
					m.Cmdy(mdb.EXPORT, TASK, "", mdb.ZONE)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(kit.MDB_ZONE)
					m.Cmdy(mdb.IMPORT, TASK, "", mdb.ZONE)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					_task_inputs(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				}},
				mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == TASK || arg[0] == kit.MDB_FOREACH {
						_task_search(m, arg[0], arg[1], arg[2])
						m.PushPodCmd(TASK, kit.Simple(mdb.SEARCH, arg)...)
					}
				}},

				BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, PROCESS)
				}},
				END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, FINISH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,zone,count", m.Conf(TASK, kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, TASK, "", mdb.ZONE, arg); len(arg) == 0 {
					m.PushAction(mdb.REMOVE)
					return
				}
				status := map[string]int{}
				m.Table(func(index int, value map[string]string, head []string) {
					m.PushButton(_task_action(m, value[STATUS]))
					status[value[kit.MDB_STATUS]]++
				})
				m.Status(status)
			}},
		},
	})
}
