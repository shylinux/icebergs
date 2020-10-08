package team

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"time"
)

const PLAN = "plan"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			PLAN: {Name: "plan scale=day,week,month,year,long begin_time@date auto 添加 导出 导入", Help: "计划", Meta: kit.Dict(
				"display", "/plugin/local/team/plan.js", "style", "plan",
			), Action: map[string]*ice.Action{
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

				gdb.BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), TaskField.STATUS, TaskStatus.PROCESS)
				}},
				gdb.END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), TaskField.STATUS, TaskStatus.FINISH)
				}},

				"command": {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) == 1 {
						m.Cmdy(ctx.COMMAND, arg[0])
						return
					}
					m.Cmdy(arg[0], arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				begin_time, end_time := _task_scope(m, 8, arg...)
				m.Set(ice.MSG_OPTION, "begin_time")
				m.Set(ice.MSG_OPTION, "end_time")

				m.Option(mdb.FIELDS, "begin_time,close_time,zone,id,level,status,score,type,name,text,extra")
				m.Option("cache.cb", func(key string, fields []string, value, val map[string]interface{}) {
					begin, _ := time.ParseInLocation(ice.MOD_TIME, kit.Format(value[TaskField.BEGIN_TIME]), time.Local)
					if begin_time.After(begin) || begin.After(end_time) {
						return
					}
					m.Push(key, value, fields, val)
					m.PushRender(kit.MDB_ACTION, kit.MDB_BUTTON, _task_action(m, value[TaskField.STATUS], "插件"))
				})
				m.Cmdy(mdb.SELECT, m.Prefix(TASK), kit.Keys(m.Option(ice.MSG_DOMAIN)), mdb.ZONE, kit.MDB_FOREACH)
			}},
		},
	}, nil)
}
