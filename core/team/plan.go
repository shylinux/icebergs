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
			PLAN: {Name: "plan scale=day,week,month,year,long begin_time@date auto insert export import", Help: "计划", Meta: kit.Dict(
				"display", "/plugin/local/team/plan.js", "style", "plan",
			), Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_task_create(m, arg[1])
					_task_insert(m, arg[1], arg[2:]...)
					m.Set(ice.MSG_RESULT)
					m.Cmdy(m.Prefix(PLAN), m.Option("scale"))
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1])
					m.Set(ice.MSG_RESULT)
					m.Cmdy(m.Prefix(PLAN), m.Option("scale"))
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					_task_delete(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID))
					m.Set(ice.MSG_RESULT)
					m.Cmdy(m.Prefix(PLAN), m.Option("scale"))
				}},
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					_task_export(m, m.Option(kit.MDB_FILE))
					m.Set(ice.MSG_RESULT)
					m.Cmdy(m.Prefix(PLAN), m.Option("scale"))
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_task_import(m, m.Option(kit.MDB_FILE))
					m.Set(ice.MSG_RESULT)
					m.Cmdy(m.Prefix(PLAN), m.Option("scale"))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "pod", "ctx", "cmd", "arg":

					default:
						_task_inputs(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
					}
				}},

				gdb.BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), TaskField.STATUS, TaskStatus.PROCESS)
				}},
				gdb.END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), TaskField.STATUS, TaskStatus.FINISH)
				}},

				mdb.PLUGIN: {Name: "plugin extra.pod extra.ctx extra.cmd extra.arg", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), kit.MDB_TIME, m.Time(),
						kit.Simple(kit.Dict(arg))...)
				}},
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == "run" {
						m.Cmdy(arg[1], arg[2:])
						return
					}
					if len(arg) == 1 {
						m.Cmdy(ctx.COMMAND, arg[0])
						return
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				begin_time, end_time := _task_scope(m, 8, arg...)
				m.Set(ice.MSG_OPTION, "begin_time")
				m.Set(ice.MSG_OPTION, "end_time")

				m.Option(mdb.FIELDS, "begin_time,close_time,zone,id,level,status,score,type,name,text,extra")
				m.Option(mdb.SELECT_CB, func(key string, fields []string, value, val map[string]interface{}) {
					begin, _ := time.ParseInLocation(ice.MOD_TIME, kit.Format(value[TaskField.BEGIN_TIME]), time.Local)
					if begin_time.After(begin) || begin.After(end_time) {
						return
					}
					m.Push(key, value, fields, val)
					m.PushButton(_task_action(m, value[TaskField.STATUS], mdb.PLUGIN))
				})
				m.Cmdy(mdb.SELECT, m.Prefix(TASK), "", mdb.ZONE, kit.MDB_FOREACH)
			}},
		},
	}, nil)
}
