package team

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"time"
)

const (
	ZONE = "zone"
	TASK = "task"
	PLAN = "plan"
)

const (
	StatusPrepare = "prepare"
	StatusProcess = "process"
	StatusCancel  = "cancel"
	StatusFinish  = "finish"
)

func _task_list(m *ice.Message, zone string, id string, field ...interface{}) {
	m.Richs(TASK, nil, kit.Select("*", zone), func(key string, val map[string]interface{}) {
		if zone = kit.Format(kit.Value(val, "meta.zone")); id == "" {
			m.Grows(TASK, kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
				m.Push(zone, value, []string{"begin_time"})
				m.Push("zone", zone)
				m.Push(zone, value, []string{"id", "status", "type", "name", "text"})
			})
			return
		}
		m.Grows(TASK, kit.Keys("hash", key), "id", id, func(index int, value map[string]interface{}) {
			m.Push("detail", value)
		})
	})
}
func _task_create(m *ice.Message, zone string) {
	m.Rich(TASK, nil, kit.Data(ZONE, zone))
	m.Log_CREATE("zone", zone)
}
func _task_insert(m *ice.Message, zone, kind, name, text, begin_time, close_time string, arg ...string) {
	m.Richs(TASK, nil, zone, func(key string, value map[string]interface{}) {
		id := m.Grow(TASK, kit.Keys("hash", key), kit.Dict(
			kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
			"begin_time", begin_time, "close_time", begin_time, "status", StatusPrepare,
			kit.MDB_EXTRA, kit.Dict(arg),
		))
		m.Log_INSERT("zone", zone, "id", id, "type", kind, "name", name)
		m.Echo("%d", id)
	})
}
func _task_modify(m *ice.Message, zone, id, pro, set, old string) {
	m.Richs(TASK, nil, kit.Select("*", zone), func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys("hash", key), "id", id, func(index int, value map[string]interface{}) {
			switch key {
			case "id", "time":
				m.Info("not allow %v", key)
			default:
				m.Log_MODIFY("zone", zone, "id", id, "key", pro, "value", set, "old", old)
				kit.Value(value, pro, set)
			}
		})
	})
}
func _task_delete(m *ice.Message, zone, id string) {
	m.Richs(TASK, nil, kit.Select("*", zone), func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys("hash", key), "id", id, func(index int, value map[string]interface{}) {
			m.Log_DELETE("zone", zone, "id", id)
			kit.Value(value, "status", "cancel")
		})
	})
}
func _task_remove(m *ice.Message, zone string) {
}

var Index = &ice.Context{Name: "team", Help: "团队中心",
	Configs: map[string]*ice.Config{
		TASK: {Name: "task", Help: "task", Value: kit.Data(kit.MDB_SHORT, "zone")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save("task") }},

		TASK: {Name: "task zone=auto id=auto auto", Help: "任务", Action: map[string]*ice.Action{
			"modify": {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, m.Option("zone"), m.Option("id"), arg[0], arg[1], arg[2])
			}},
			"delete": {Name: "delete key value", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				_task_delete(m, m.Option("zone"), m.Option("id"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 3 {
				_task_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				return
			}

			if m.Richs(TASK, nil, arg[0], nil) == nil {
				_task_create(m, arg[0])
			}
			if len(arg) == 5 {
				arg = append(arg, m.Time())
			}
			if len(arg) == 6 {
				arg = append(arg, m.Time("1h"))
			}

			_task_insert(m, arg[0], arg[2], arg[3], arg[4], arg[5], arg[6], arg[7:]...)
		}},
		PLAN: {Name: "plan scale:select=day|week|month|year begin_time=@date end_time=@date auto", Help: "计划", Meta: kit.Dict(
			"display", "/plugin/local/team/miss.js", "detail", []string{"prepare", "process", "finish", "cancel"},
		), Action: map[string]*ice.Action{
			"insert": {Name: "insert begin_time type name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_task_insert(m, arg[0], arg[1], arg[2], arg[3], arg[4], arg[5])
			}},
			"delete": {Name: "delete key value", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				_task_delete(m, m.Option("zone"), m.Option("id"))
			}},
			"modify": {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, m.Option("zone"), m.Option("id"), arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			begin_time := time.Now()
			if len(arg) > 1 {
				begin_time, _ = time.ParseInLocation("2006-01-02 15:04:05", arg[1], time.Local)
			}

			end_time := begin_time
			if len(arg) > 2 {
				end_time, _ = time.ParseInLocation("2006-01-02 15:04:05", arg[2], time.Local)
			}

			switch arg[0] {
			case "day":
				begin_time = begin_time.Add(-time.Duration(begin_time.UnixNano()) % (24 * time.Hour))
				end_time = begin_time.Add(24 * time.Hour)
			case "week":
				begin_time = begin_time.Add(-time.Duration(begin_time.UnixNano())%(24*time.Hour) - time.Duration(begin_time.Weekday())*24*time.Hour)
				end_time = begin_time.Add(7 * 24 * time.Hour)
			case "month":
			case "months":
			case "year":
			case "long":
			}

			begin_time = begin_time.Add(-8 * time.Hour)
			end_time = end_time.Add(-8 * time.Hour)
			m.Debug("begin_time: %v %v", begin_time, end_time)

			m.Richs(TASK, nil, kit.Select("*", m.Option("zone")), func(key string, val map[string]interface{}) {
				zone := kit.Format(kit.Value(val, "meta.zone"))
				m.Grows(TASK, kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
					begin, _ := time.ParseInLocation("2006-01-02 15:04:05", kit.Format(value["begin_time"]), time.Local)
					m.Debug("begin: %v", begin)

					if begin_time.Before(begin) && begin.Before(end_time) {
						m.Push(zone, value, []string{"begin_time"})
						m.Push("zone", zone)
						m.Push(zone, value, []string{"id", "status", "type", "name", "text"})
					}
				})
			})
		}},
		"miss": {Name: "miss", Help: "miss"},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
