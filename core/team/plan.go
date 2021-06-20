package team

import (
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _plan_list(m *ice.Message, begin_time, end_time time.Time) *ice.Message {
	m.Option(mdb.CACHE_LIMIT, "100")
	m.Fields(true, "begin_time,close_time,zone,id,level,status,score,type,name,text,extra")
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
			PLAN: {Name: "plan scale=day,week,month,year,long begin_time@date auto insert export import", Help: "计划", Meta: kit.Dict(
				kit.MDB_DISPLAY, "/plugin/local/team/plan.js", kit.MDB_STYLE, PLAN,
			), Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_task_create(m, arg[1])
					_task_insert(m, arg[1], arg[2:]...)
					m.ProcessRefresh("30ms")
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1])
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					_task_export(m, "")
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_task_import(m, "")
					m.ProcessRefresh("30ms")
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					_task_inputs(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				}},

				mdb.PLUGIN: {Name: "plugin extra.ctx extra.cmd extra.arg", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), kit.MDB_TIME, m.Time(), arg...)
					m.Set(ice.MSG_RESULT).Cmdy(PLAN, m.Option(SCALE))
				}},
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == cli.RUN {
						m.Cmdy(arg[1], arg[2:])
						return
					}
					if len(arg) > 0 {
						m.Cmdy(ctx.COMMAND, arg[0])
					}
				}},

				BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, PROCESS)
				}},
				END: {Name: "end", Help: "结束", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, FINISH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				begin_time, end_time := _task_scope(m, 8, arg...)
				_plan_list(m, begin_time, end_time)
				m.PushPodCmd(PLAN, arg...)
			}},
		},
	})
}
