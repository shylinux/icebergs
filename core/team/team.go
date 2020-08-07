package team

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"encoding/csv"
	"os"
	"path"
	"strings"
	"time"
)

func _task_list(m *ice.Message, zone string, id string, field ...interface{}) {
	fields := strings.Split(kit.Select("begin_time,zone,id,status,level,type,name,text", m.Option("fields")), ",")
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		if zone = kit.Format(kit.Value(val, "meta.zone")); id == "" {
			m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				m.Push(zone, value, fields)
			})
			return
		}
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Push("detail", value)
		})
	})
}
func _task_create(m *ice.Message, zone string) {
	if m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), zone, nil) == nil {
		m.Conf(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_META, kit.MDB_SHORT), kit.MDB_ZONE)
		m.Rich(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Data(kit.MDB_ZONE, zone))
		m.Log_CREATE(kit.MDB_ZONE, zone)
	}
}
func _task_insert(m *ice.Message, zone string, arg ...string) {
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), zone, func(key string, value map[string]interface{}) {
		id := m.Grow(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), kit.Dict(
			BEGIN_TIME, m.Time(), CLOSE_TIME, m.Time(), kit.MDB_EXTRA, kit.Dict(),
			STATUS, StatusPrepare, LEVEL, 3, SCORE, 3, arg,
		))
		m.Log_INSERT(kit.MDB_ZONE, zone, kit.MDB_ID, id, arg[0], arg[1])
		m.Echo("%d", id)
	})
}
func _task_modify(m *ice.Message, zone, id, pro, set string) {
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			switch pro {
			case kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TIME:
				m.Info("not allow %v", key)
			case STATUS:
				if value[STATUS] == set {
					break
				}
				switch value[STATUS] {
				case StatusCancel, StatusFinish:
					m.Info("not allow %v", key)
					return
				}
				switch set {
				case StatusProcess:
					kit.Value(value, BEGIN_TIME, m.Time())
				case StatusCancel, StatusFinish:
					kit.Value(value, CLOSE_TIME, m.Time())
				}
				fallthrough
			default:
				m.Log_MODIFY(kit.MDB_ZONE, zone, kit.MDB_ID, id, kit.MDB_KEY, pro, kit.MDB_VALUE, set)
				kit.Value(value, pro, set)
			}
		})
	})
}
func _task_delete(m *ice.Message, zone, id string) {
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Log_DELETE(kit.MDB_ZONE, zone, kit.MDB_ID, id)
			kit.Value(value, STATUS, StatusCancel)
		})
	})
}
func _task_export(m *ice.Message, file string) {
	f, p, e := kit.Create(file)
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	m.Assert(w.Write([]string{
		kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TIME,
		kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT,
		LEVEL, STATUS, SCORE,
		BEGIN_TIME, CLOSE_TIME,
		kit.MDB_EXTRA,
	}))
	count := 0
	m.Option("cache.limit", -2)
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			m.Assert(w.Write(kit.Simple(
				kit.Format(kit.Value(val, "meta.zone")),
				kit.Format(value[kit.MDB_ID]),
				kit.Format(value[kit.MDB_TIME]),
				kit.Format(value[kit.MDB_TYPE]),
				kit.Format(value[kit.MDB_NAME]),
				kit.Format(value[kit.MDB_TEXT]),
				kit.Format(value[LEVEL]),
				kit.Format(value[STATUS]),
				kit.Format(value[SCORE]),
				kit.Format(value[BEGIN_TIME]),
				kit.Format(value[CLOSE_TIME]),
				kit.Format(value[kit.MDB_EXTRA]),
			)))
			count++
		})
	})
	m.Log_EXPORT("file", p, "count", count)
	m.Echo(p)
}
func _task_import(m *ice.Message, file string) {
	f, e := os.Open(file)
	m.Assert(e)
	defer f.Close()

	r := csv.NewReader(f)
	heads, _ := r.Read()
	count := 0
	for {
		lines, e := r.Read()
		if e != nil {
			break
		}

		zone := ""
		data := kit.Dict()
		for i, k := range heads {
			switch k {
			case kit.MDB_ZONE:
				zone = lines[i]
			case kit.MDB_ID:
				continue
			case kit.MDB_EXTRA:
				kit.Value(data, k, kit.UnMarshal(lines[i]))
			default:
				kit.Value(data, k, lines[i])
			}
		}

		_task_create(m, zone)
		m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), zone, func(key string, value map[string]interface{}) {
			id := m.Grow(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), data)
			m.Log_INSERT(kit.MDB_ZONE, zone, kit.MDB_ID, id)
			count++
		})
	}
	m.Log_IMPORT("file", file, "count", count)
	m.Echo(file)
}

func _task_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			if value[kit.MDB_NAME] == name || strings.Contains(kit.Format(value[kit.MDB_TEXT]), name) {
				m.Push("pod", m.Option(ice.MSG_USERPOD))
				m.Push("ctx", m.Prefix())
				m.Push("cmd", TASK)
				m.Push("time", value[kit.MDB_TIME])
				m.Push("size", 1)
				m.Push("type", TASK)
				m.Push("name", value[kit.MDB_NAME])
				m.Push("text", kit.Format("%s:%d", kit.Value(val, "meta.zone"), kit.Int(value[kit.MDB_ID])))

			}
		})
	})
}
func _task_render(m *ice.Message, kind, name, text string, arg ...string) {
	ls := strings.Split(text, ":")
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), ls[0], func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "id", ls[1], func(index int, value map[string]interface{}) {
			m.Push("detail", value)
		})
	})
}
func _task_action(m *ice.Message, status interface{}, action ...string) string {
	switch status {
	case StatusPrepare:
		action = append(action, "开始")
	case StatusProcess:
		action = append(action, "完成")
	case StatusCancel, StatusFinish:
	}
	for i, v := range action {
		action[i] = m.Cmdx(mdb.RENDER, web.RENDER.Button, v)
	}
	return strings.Join(action, "")
}
func _task_input(m *ice.Message, field, value string) {
	switch field {
	case "zone":
		m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
			m.Push("zone", kit.Value(val, "meta.zone"))
			m.Push("count", kit.Select("0", kit.Format(kit.Value(val, "meta.count"))))
		})
		m.Sort("count", "int_r")
	case "name", "text":
		list := map[string]int{}
		m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
			m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				list[kit.Format(value[field])]++
			})
		})
		for k, i := range list {
			m.Push("key", k)
			m.Push("count", i)
		}
		m.Sort("count", "int_r")
	}
}
func _task_scope(m *ice.Message, arg ...string) (time.Time, time.Time) {
	begin_time := time.Now()
	if len(arg) > 1 {
		begin_time, _ = time.ParseInLocation(ice.MOD_TIME, arg[1], time.Local)
	}
	end_time := begin_time

	switch begin_time = begin_time.Add(-time.Duration(begin_time.UnixNano()) % (24 * time.Hour)); arg[0] {
	case ScaleDay:
		end_time = begin_time.AddDate(0, 0, 1)
	case ScaleWeek:
		begin_time = begin_time.AddDate(0, 0, -int(begin_time.Weekday()))
		end_time = begin_time.AddDate(0, 0, 7)
	case ScaleMonth:
		begin_time = begin_time.AddDate(0, 0, -begin_time.Day()+1)
		end_time = begin_time.AddDate(0, 1, 0)
	case ScaleYear:
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1)
		end_time = begin_time.AddDate(1, 0, 0)
	case ScaleLong:
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1)
		begin_time = begin_time.AddDate(-begin_time.Year()%5, 0, 0)
		end_time = begin_time.AddDate(5, 0, 0)
	}

	begin_time = begin_time.Add(-8 * time.Hour)
	end_time = end_time.Add(-8 * time.Hour)
	return begin_time, end_time
}

var _task_inputs = kit.List(
	"_input", "text", "name", "zone", "value", "@key",
	"_input", "select", "name", "type", "values", []interface{}{
		"once", "step", "week",
	},
	"_input", "text", "name", "name", "value", "@key",
	"_input", "text", "name", "text", "value", "@key",
	"_input", "text", "name", "extra.cmds",
	"_input", "text", "name", "extra.args",
	"_input", "text", "name", "begin_time", "value", "@date",
	"_input", "text", "name", "close_time", "value", "@date",
)

const (
	TASK = "task"
	PLAN = "plan"
	MISS = "miss"
)
const (
	LEVEL  = "level"
	STATUS = "status"
	SCORE  = "score"

	BEGIN_TIME = "begin_time"
	CLOSE_TIME = "close_time"

	EXPORT = "usr/export/web.team.task/"
)
const (
	ScaleDay   = "day"
	ScaleWeek  = "week"
	ScaleMonth = "month"
	ScaleYear  = "year"
	ScaleLong  = "long"
)
const (
	StatusPrepare = "prepare"
	StatusProcess = "process"
	StatusCancel  = "cancel"
	StatusFinish  = "finish"
)

var Index = &ice.Context{Name: "team", Help: "团队中心",
	Configs: map[string]*ice.Config{
		TASK: {Name: "task", Help: "task", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
		MISS: {Name: "miss", Help: "miss", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			m.Cmd(mdb.SEARCH, mdb.CREATE, TASK, TASK, m.Prefix())
			m.Cmd(mdb.RENDER, mdb.CREATE, TASK, TASK, m.Prefix())
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(TASK) }},

		TASK: {Name: "task zone=auto id=auto auto 添加:button 导出:button 导入:button", Help: "任务", Meta: kit.Dict(
			"添加", _task_inputs,
		), Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert [key value]...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_task_create(m, arg[1])
				_task_insert(m, arg[1], arg[2:]...)
			}},
			mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1])
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				_task_delete(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID))
			}},
			mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				_task_export(m, kit.Select(path.Join(EXPORT, m.Option(ice.MSG_DOMAIN), "list.csv"), arg, 0))
			}},
			mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				_task_import(m, kit.Select(path.Join(EXPORT, m.Option(ice.MSG_DOMAIN), "list.csv"), arg, 0))
			}},

			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_task_search(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_task_render(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},

			gdb.BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, StatusProcess)
			}},
			gdb.END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, StatusFinish)
			}},

			"input": {Name: "input key value", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				_task_input(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if _task_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1)); len(arg) < 2 {
				m.Table(func(index int, value map[string]string, head []string) {
					m.Push("action", _task_action(m, value[STATUS]))
				})
			} else {
				m.Table(func(index int, value map[string]string, head []string) {
					if value["key"] == "status" {
						m.Push("key", "action")
						m.Push("value", _task_action(m, value["value"]))
					}
				})
			}
		}},
		PLAN: {Name: "plan scale:select=day|week|month|year|long begin_time=@date auto 添加:button 导出:button 导入:button 筛选:button", Help: "计划", Meta: kit.Dict(
			"display", "/plugin/local/team/plan.js", "style", "plan",
			"添加", _task_inputs,
		), Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert [key value]...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_task_create(m, arg[1])
				_task_insert(m, arg[1], arg[2:]...)
			}},
			mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
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
			"input": {Name: "input key value", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				_task_input(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			begin_time, end_time := _task_scope(m, arg...)
			m.Logs("info", "begin", begin_time, "end", end_time)

			m.Set(ice.MSG_OPTION, "end_time")
			m.Set(ice.MSG_OPTION, "begin_time")
			m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Select(kit.MDB_FOREACH, m.Option(kit.MDB_ZONE)), func(key string, val map[string]interface{}) {
				zone := kit.Format(kit.Value(val, "meta.zone"))
				m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
					begin, _ := time.ParseInLocation(ice.MOD_TIME, kit.Format(value[BEGIN_TIME]), time.Local)
					if begin_time.Before(begin) && begin.Before(end_time) {
						m.Push(zone, value)
						m.Push(kit.MDB_ZONE, zone)
						m.Push("action", _task_action(m, value[STATUS], "插件"))
					}
				})
			})
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
