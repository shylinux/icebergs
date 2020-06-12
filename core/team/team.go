package team

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"encoding/csv"
	"os"
	"time"
)

const (
	PLAN = "plan"
	TASK = "task"
	MISS = "miss"
)
const (
	ZONE   = "zone"
	LEVEL  = "level"
	STATUS = "status"
	SCORE  = "score"

	BEGIN_TIME = "begin_time"
	CLOSE_TIME = "close_time"

	EXPORT = "usr/export/web.team/task.csv"
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

func _task_list(m *ice.Message, zone string, id string, field ...interface{}) {
	m.Richs(TASK, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		if zone = kit.Format(kit.Value(val, "meta.zone")); id == "" {
			m.Grows(TASK, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				m.Push(zone, value)
				m.Push(ZONE, zone)
			})
			return
		}
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Push("detail", value)
		})
	})
}
func _task_create(m *ice.Message, zone string) {
	if m.Richs(TASK, nil, zone, nil) == nil {
		m.Rich(TASK, nil, kit.Data(ZONE, zone))
		m.Log_CREATE(ZONE, zone)
	}
}
func _task_export(m *ice.Message, file string) {
	f, p, e := kit.Create(file)
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	m.Assert(w.Write([]string{
		ZONE, kit.MDB_ID, kit.MDB_TIME,
		kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT,
		LEVEL, STATUS, SCORE,
		BEGIN_TIME, CLOSE_TIME,
		kit.MDB_EXTRA,
	}))
	count := 0
	m.Option("cache.limit", -2)
	m.Richs(TASK, nil, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
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
			case ZONE:
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
		m.Richs(TASK, nil, zone, func(key string, value map[string]interface{}) {
			id := m.Grow(TASK, kit.Keys(kit.MDB_HASH, key), data)
			m.Log_INSERT(ZONE, zone, kit.MDB_ID, id)
			count++
		})
	}
	m.Log_IMPORT("file", file, "count", count)
}

func _task_insert(m *ice.Message, zone, kind, name, text, begin_time, close_time string, arg ...string) {
	m.Richs(TASK, nil, zone, func(key string, value map[string]interface{}) {
		id := m.Grow(TASK, kit.Keys(kit.MDB_HASH, key), kit.Dict(
			kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
			BEGIN_TIME, begin_time, CLOSE_TIME, begin_time,
			STATUS, StatusPrepare, LEVEL, 3, SCORE, 3,
			kit.MDB_EXTRA, kit.Dict(arg),
		))
		m.Log_INSERT(ZONE, zone, kit.MDB_ID, id, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
		m.Echo("%d", id)
	})
}
func _task_modify(m *ice.Message, zone, id, pro, set, old string) {
	m.Richs(TASK, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			switch pro {
			case ZONE, kit.MDB_ID, kit.MDB_TIME:
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
				m.Log_MODIFY(ZONE, zone, kit.MDB_ID, id, kit.MDB_KEY, pro, kit.MDB_VALUE, set, "old", old)
				kit.Value(value, pro, set)
			}
		})
	})
}
func _task_delete(m *ice.Message, zone, id string) {
	m.Richs(TASK, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Log_DELETE(ZONE, zone, kit.MDB_ID, id)
			kit.Value(value, STATUS, StatusCancel)
		})
	})
}
func _task_plugin(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		kit.Fetch(m.Confv(MISS, "meta.plug"), func(key string, value map[string]interface{}) {
			for k := range value {
				m.Push(key, k)
			}
		})
		return
	}
	m.Cmdy(kit.Keys(arg[0], arg[1]), arg[2:])
}
func _task_input(m *ice.Message, key, value string) {
	m.Cmdy(kit.Keys(m.Option("zone"), m.Option("type")), "action", "input", key, value)
}

var Index = &ice.Context{Name: "team", Help: "团队中心",
	Configs: map[string]*ice.Config{
		TASK: {Name: "task", Help: "task", Value: kit.Data(kit.MDB_SHORT, ZONE)},
		MISS: {Name: "miss", Help: "miss", Value: kit.Data(kit.MDB_SHORT, ZONE)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
				if s == c {
					return
				}
				m.Conf(MISS, kit.Keys("meta.plug", s.Name, key), cmd.Name)
			})
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(TASK) }},

		PLAN: {Name: "plan scale:select=day|week|month|year|long begin_time=@date end_time=@date auto", Help: "计划", Meta: kit.Dict(
			"display", "/plugin/local/team/plan.js", "detail", []string{StatusPrepare, StatusProcess, StatusCancel, StatusFinish},
		), Action: map[string]*ice.Action{
			kit.MDB_INSERT: {Name: "insert zone type name text begin_time end_time", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_task_create(m, arg[0])
				_task_insert(m, arg[0], arg[1], arg[2], arg[3], arg[4], arg[5])
			}},
			kit.MDB_DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				_task_delete(m, m.Option(ZONE), m.Option(kit.MDB_ID))
			}},
			kit.MDB_MODIFY: {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, m.Option(ZONE), m.Option(kit.MDB_ID), arg[0], arg[1], arg[2])
			}},
			kit.MDB_IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				_task_import(m, kit.Select(EXPORT, arg, 0))
			}},
			kit.MDB_EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				_task_export(m, kit.Select(EXPORT, arg, 0))
			}},
			"plugin": {Name: "plugin", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
				_task_plugin(m, arg...)
			}},
			"input": {Name: "input", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				_task_input(m, arg[0], arg[1])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			begin_time := time.Now()
			if len(arg) > 1 {
				begin_time, _ = time.ParseInLocation(ice.ICE_TIME, arg[1], time.Local)
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
			m.Logs("info", "begin", begin_time, "end", end_time)

			m.Richs(TASK, nil, kit.Select(kit.MDB_FOREACH, m.Option(ZONE)), func(key string, val map[string]interface{}) {
				zone := kit.Format(kit.Value(val, "meta.zone"))
				m.Grows(TASK, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
					begin, _ := time.ParseInLocation(ice.ICE_TIME, kit.Format(value[BEGIN_TIME]), time.Local)
					if begin_time.Before(begin) && begin.Before(end_time) {
						m.Push(zone, value)
						m.Push(ZONE, zone)
					}
				})
			})
		}},
		TASK: {Name: "task zone=auto id=auto auto", Help: "任务", Action: map[string]*ice.Action{
			kit.MDB_DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				_task_delete(m, m.Option(ZONE), m.Option(kit.MDB_ID))
			}},
			kit.MDB_MODIFY: {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				_task_modify(m, m.Option(ZONE), m.Option(kit.MDB_ID), arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 3 {
				_task_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				return
			}

			if len(arg) == 5 {
				arg = append(arg, m.Time())
			}
			if len(arg) == 6 {
				arg = append(arg, m.Time("1h"))
			}

			_task_create(m, arg[0])
			_task_insert(m, arg[0], arg[2], arg[3], arg[4], arg[5], arg[6], arg[7:]...)
		}},
		MISS: {Name: "miss", Help: "miss", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
