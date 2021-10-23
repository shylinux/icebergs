package team

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _plan_list(m *ice.Message, begin_time, end_time time.Time) *ice.Message {
	m.Option(ice.CACHE_LIMIT, "100")
	m.Fields(0, "begin_time,close_time,zone,id,level,status,score,type,name,text,extra")
	m.Option(kit.Keycb(mdb.SELECT), func(key string, fields []string, value, val map[string]interface{}) {
		begin, _ := time.ParseInLocation(ice.MOD_TIME, kit.Format(value[BEGIN_TIME]), time.Local)
		if begin_time.After(begin) || begin.After(end_time) {
			return
		}
		m.Push(key, value, fields, val)
		m.PushButton(_task_action(m, value[STATUS], mdb.PLUGIN))
	})
	m.Cmd(mdb.SELECT, TASK, "", mdb.ZONE, kit.MDB_FOREACH)
	return m
}

const (
	BEGIN = "begin"
	END   = "end"
)

const PLAN = "plan"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			PLAN: {Name: "plan scale=week,day,week,month,year,long begin_time@date place@province auto insert export import", Help: "计划", Meta: kit.Dict(
				ice.Display("/plugin/local/team/plan.js", PLAN),
			), Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(TASK, mdb.INSERT, arg)
					m.ProcessRefresh30ms()
				}},
				mdb.MODIFY: {Name: "task modify", Help: "编辑"},
				mdb.EXPORT: {Name: "task export", Help: "导出"},
				mdb.IMPORT: {Name: "task import", Help: "导入"},
				mdb.INPUTS: {Name: "task inputs", Help: "补全"},

				mdb.PLUGIN: {Name: "plugin extra.ctx extra.cmd extra.arg", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, arg[0], arg[1], arg[2:]...)
					m.ProcessRefresh30ms()
				}},
				ctx.COMMAND: {Name: "command", Help: "命令"},
				ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},

				BEGIN: {Name: "task begin", Help: "开始"},
				END:   {Name: "task end", Help: "结束"},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				begin_time, end_time := _task_scope(m, 8, arg...)
				_plan_list(m, begin_time, end_time)
				m.PushPodCmd(PLAN, arg...)
			}},
		},
	})
}
