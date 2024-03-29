package team

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _plan_scope(m *ice.Message, arg ...string) (begin_time, end_time time.Time) {
	switch begin_time = kit.DayBegin(kit.Select(m.Time(), arg, 1)); kit.Select(WEEK, arg, 0) {
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
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1).AddDate(-5, 0, 0)
		end_time = begin_time.AddDate(10, 0, 0)
	}
	m.Logs(mdb.SELECT, "begin_time", begin_time, "end_time", end_time)
	return begin_time, end_time
}
func _plan_list(m *ice.Message, begin_time, end_time string) {
	m.Options(mdb.CACHE_LIMIT, "-1").OptionFields("begin_time,close_time,zone,id,level,status,score,type,name,text,extra")
	m.Cmd(mdb.SELECT, m.Prefix(TASK), "", mdb.ZONE, mdb.FOREACH, func(key string, fields []string, value, val ice.Map) {
		if begin_time <= kit.Format(value[BEGIN_TIME]) && kit.Format(value[BEGIN_TIME]) < end_time {
			m.Push(key, value, fields, val).PushButton(_task_action(m, value[STATUS], mdb.PLUGIN))
		}
	})
}

const (
	DAY   = "day"
	WEEK  = "week"
	MONTH = "month"
	YEAR  = "year"
	LONG  = "long"
)
const (
	TASK_POD = "task.pod"
)

const PLAN = "plan"

func init() {
	Index.MergeCommands(ice.Commands{
		PLAN: {Name: "plan scale=week,day,week,month,year,long begin_time@date list prev next", Help: "计划表", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TODO, mdb.INPUTS, arg) }},
			mdb.PLUGIN: {Name: "plugin extra.index extra.args", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TASK, mdb.MODIFY, arg) }},
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if m.RenameOption(TASK_POD, ice.POD); ctx.PodCmd(m, m.PrefixKey(), ice.RUN, arg) {
					return
				}
				if cmd := m.CmdAppend(TASK, kit.Slice(arg, 0, 2), ctx.INDEX); cmd != "" {
					m.Cmdy(cmd, arg[2:])
				} else if aaa.Right(m, arg) {
					m.Cmdy(arg)
				}
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() { m.PushButton(kit.Dict(m.CommandKey(), "计划")) })
			}},
		}, aaa.RoleAction(ctx.COMMAND), ctx.CmdAction(), TASK), Hand: func(m *ice.Message, arg ...string) {
			begin_time, end_time := _plan_scope(m, kit.Slice(arg, 0, 2)...)
			_plan_list(m, begin_time.Format(ice.MOD_TIME), end_time.Format(ice.MOD_TIME))
			web.PushPodCmd(m, "", arg...)
			ctx.DisplayLocal(m, "")
			ctx.Toolkit(m, TODO, TASK, EPIC)
		}},
	})
}
