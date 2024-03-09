package team

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
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
	m.Logs(mdb.SELECT, BEGIN_TIME, begin_time, END_TIME, end_time)
	return begin_time, end_time
}
func _plan_list(m *ice.Message, begin_time, end_time string) {
	m.Options(mdb.CACHE_LIMIT, "-1").OptionFields("begin_time,end_time,zone,id,status,level,score,type,name,text,extra")
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
	SCALE    = "scale"
)

const PLAN = "plan"

func init() {
	Index.MergeCommands(ice.Commands{
		PLAN: {Name: "plan scale=month,day,week,month,year,long begin_time@date list insert prev next actions", Help: "计划表", Icon: "Calendar.png", Role: aaa.VOID, Meta: kit.Dict(
			ctx.TRANS, kit.Dict(html.INPUT, kit.Dict(
				SCALE, "跨度", "view", "视图",
				DAY, "日", WEEK, "周", MONTH, "月", YEAR, "年", LONG, "代",
			)),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TODO, mdb.INPUTS, arg) }},
			mdb.PLUGIN: {Name: "plugin extra.index extra.args", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TASK, mdb.MODIFY, arg) }},
			mdb.INSERT: {Name: "insert space zone* type*=once,step,week name* text begin_time@date end_time@date", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TASK, mdb.INSERT, arg)
				web.StreamPushRefreshConfirm(m, m.Trans("refresh for new message ", "刷新列表，查看最新消息 "))
			}},
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				if ice.Info.Important {
					PlanInsertPlan(m, web.DREAM, "", m.Option(mdb.NAME), web.CHAT_IFRAME, web.S(m.Option(mdb.NAME)))
				}
			}},
			web.DREAM_REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				PlanInsertPlan(m, web.DREAM, "", "", web.CHAT_IFRAME, web.S(m.Option(mdb.NAME)))
			}},
			aaa.OFFER_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				PlanInsertPlan(m, aaa.APPLY, "", m.Option(aaa.EMAIL), aaa.OFFER, m.Option(mdb.HASH))
			}},
			aaa.OFFER_ACCEPT: {Hand: func(m *ice.Message, arg ...string) {
				PlanInsertPlan(m, aaa.APPLY, "", m.Option(aaa.EMAIL), aaa.OFFER, m.Option(mdb.HASH))
			}},
			aaa.USER_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				PlanInsertPlan(m, aaa.APPLY, "", "", aaa.USER, m.Option(aaa.USERNAME))
			}},
			aaa.USER_REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				PlanInsertPlan(m, aaa.APPLY, "", "", aaa.USER, m.Option(aaa.USERNAME))
			}},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if m.RenameOption(TASK_POD, ice.POD); ctx.PodCmd(m, m.ShortKey(), ctx.RUN, arg) {
					return
				} else if cmd := m.CmdAppend(TASK, kit.Slice(arg, 0, 2), ctx.INDEX); cmd != "" {
					m.Cmdy(cmd, arg[2:])
				} else if aaa.Right(m, arg) {
					m.Cmdy(arg)
				}
			}},
		}, web.DreamTablesAction(), web.DreamAction(), aaa.OfferAction(), ctx.ConfAction(mdb.TOOLS, kit.Simple(TODO, EPIC), web.ONLINE, ice.TRUE), TASK), Hand: func(m *ice.Message, arg ...string) {
			begin_time, end_time := _plan_scope(m, kit.Slice(arg, 0, 2)...)
			_plan_list(m, begin_time.Format(ice.MOD_TIME), end_time.Format(ice.MOD_TIME))
			web.PushPodCmd(m, "", arg...)
			ctx.DisplayLocal(m, "")
			ctx.Toolkit(m, "")
		}},
	})
}
func PlanInsertPlan(m *ice.Message, zone, name, text, index, args string, arg ...string) {
	if ice.Info.Important {
		m.Cmd(PLAN, mdb.INSERT, web.SPACE, "", mdb.ZONE, zone, mdb.TYPE, "once",
			mdb.NAME, kit.Select(m.ActionKey(), name), mdb.TEXT, kit.Select(args, text), BEGIN_TIME, m.Time(),
			"extra.index", kit.Select(m.ShortKey(), index), "extra.args", args, arg,
		)
	}
}
