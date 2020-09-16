package team

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
					_task_export(m, kit.Select(path.Join(EXPORT, m.Option(ice.MSG_DOMAIN), "list.csv")))
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_task_import(m, kit.Select(path.Join(EXPORT, m.Option(ice.MSG_DOMAIN), "list.csv")))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					_task_input(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				}},

				gdb.BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, StatusProcess)
				}},
				gdb.END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, StatusFinish)
				}},

				"command": {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) == 1 {
						m.Cmdy(ctx.COMMAND, arg[0])
						return
					}
					m.Cmdy(arg[0], arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				begin_time, end_time := _task_scope(m, arg...)

				m.Set(ice.MSG_OPTION, "end_time")
				m.Set(ice.MSG_OPTION, "begin_time")
				m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Select(kit.MDB_FOREACH, m.Option(kit.MDB_ZONE)), func(key string, val map[string]interface{}) {
					zone := kit.Format(kit.Value(val, "meta.zone"))
					m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
						begin, _ := time.ParseInLocation(ice.MOD_TIME, kit.Format(value[BEGIN_TIME]), time.Local)
						if begin_time.Before(begin) && begin.Before(end_time) {
							m.Push(zone, value)
							m.Push(kit.MDB_ZONE, zone)
							m.PushAction(_task_action(m, value[STATUS], "插件"))
						}
					})
				})
			}},
		},
	}, nil)
}
