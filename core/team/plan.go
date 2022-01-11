package team

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _plan_scope(m *ice.Message, tz int, arg ...string) (time.Time, time.Time) {
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
		begin_time = begin_time.AddDate(-30, 0, 0)
		end_time = begin_time.AddDate(60, 0, 0)
	}

	return begin_time, end_time
}
func _plan_list(m *ice.Message, begin_time, end_time time.Time) *ice.Message {
	m.Option(ice.CACHE_LIMIT, "100")
	m.Fields(0, "begin_time,close_time,zone,id,level,status,score,type,name,text,pod,extra")
	m.OptionCB(mdb.SELECT, func(key string, fields []string, value, val map[string]interface{}) {
		begin, _ := time.ParseInLocation(ice.MOD_TIME, kit.Format(value[BEGIN_TIME]), time.Local)
		if begin_time.After(begin) || begin.After(end_time) {
			return
		}
		m.Push(key, value, fields, val)
		m.PushButton(_task_action(m, value[STATUS], mdb.PLUGIN))
	})
	m.Cmd(mdb.SELECT, m.Prefix(TASK), "", mdb.ZONE, mdb.FOREACH)
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
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		PLAN: {Name: "plan scale=week,day,week,month,year,long begin_time@date place@province auto insert export import", Help: "计划", Meta: kit.Dict(
			ice.Display("/plugin/local/team/plan.js"),
		), Action: ice.MergeAction(map[string]*ice.Action{
			mdb.PLUGIN: {Name: "plugin extra.ctx extra.cmd extra.arg", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, arg[0], arg[1], arg[2:]...)
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.POD, m.Option("task.pod"))
				m.Option("task.pod", "")
				if m.PodCmd(m.PrefixKey(), ice.RUN, arg) {
					return
				}
				msg := m.Cmd(TASK, arg[0], arg[1])
				m.Cmdy(kit.Simple(kit.Keys(msg.Append(kit.KeyExtra(ice.CTX)), msg.Append(kit.KeyExtra(ice.CMD))), arg[2:]))
			}},
		}, TASK), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			begin_time, end_time := _plan_scope(m, 8, arg...)
			_plan_list(m, begin_time, end_time)
			m.PushPodCmd(cmd, arg...)
		}},
	}})
}
