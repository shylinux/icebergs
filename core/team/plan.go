package team

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _plan_scope(m *ice.Message, tz int, arg ...string) (begin_time, end_time time.Time) {
	if begin_time = time.Now(); len(arg) > 1 {
		begin_time, _ = time.ParseInLocation(ice.MOD_TIME, arg[1], time.Local)
	}
	begin_time = begin_time.Add(time.Duration(tz) * time.Hour)
	begin_time = begin_time.Add(-time.Duration(begin_time.UnixNano()) % (24 * time.Hour))
	begin_time = begin_time.Add(-time.Duration(tz) * time.Hour)

	switch end_time = begin_time; kit.Select(WEEK, arg, 0) {
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
		begin_time = begin_time.AddDate(-30, 0, 0)
		end_time = begin_time.AddDate(60, 0, 0)
	}
	return begin_time, end_time
}
func _plan_list(m *ice.Message, begin_time, end_time time.Time) *ice.Message {
	m.Option(mdb.CACHE_LIMIT, "-1")
	m.OptionFields("begin_time,close_time,zone,id,level,status,score,type,name,text,pod,extra")
	m.Cmd(mdb.SELECT, m.Prefix(TASK), "", mdb.ZONE, mdb.FOREACH, func(key string, fields []string, value, val ice.Map) {
		begin, _ := time.ParseInLocation(ice.MOD_TIME, kit.Format(value[BEGIN_TIME]), time.Local)
		if begin_time.After(begin) || begin.After(end_time) {
			return
		}
		m.Push(key, value, fields, val).PushButton(_task_action(m, value[STATUS], mdb.PLUGIN))
	})
	return m
}

const (
	DAY   = "day"
	WEEK  = "week"
	MONTH = "month"
	YEAR  = "year"
	LONG  = "long"
)

const PLAN = "plan"

func init() {
	Index.MergeCommands(ice.Commands{
		PLAN: {Name: "plan scale=week,day,week,month,year,long begin_time@date list", Help: "计划", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TODO, mdb.INPUTS, arg) }},
			mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TASK, mdb.INSERT, arg)
			}},
			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TASK, mdb.MODIFY, arg) }},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TASK, mdb.EXPORT) }},
			mdb.IMPORT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TASK, mdb.IMPORT) }},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.POD, m.Option("task.pod"))
				if m.Option("task.pod", ""); ctx.PodCmd(m, m.PrefixKey(), ice.RUN, arg) {
					return
				}
				m.Cmdy(m.CmdAppend(TASK, arg[0], arg[1], ctx.INDEX), arg[2:])
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmdy(TASK, arg)
				return
			}
			begin_time, end_time := _plan_scope(m, 8, kit.Slice(arg, 0, 2)...)
			_plan_list(m, begin_time, end_time)
			web.PushPodCmd(m, m.CommandKey(), arg...)
			ctx.DisplayLocal(m, "")
		}},
	})
}
