package team

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"encoding/csv"
	"strings"
	"time"
)

func DateZero(m *ice.Message, day time.Time) time.Time {
	_, z := day.Zone()
	day = day.Truncate(time.Hour * 24).Add(-time.Duration(z) * time.Second)
	return day
}

func ShowDay(m *ice.Message, day time.Time) string {
	if day.Day() == 1 {
		if day.Month() == 1 {
			return kit.Format(`<span data-year="%d" data-month="%d" data-day="%d">%d年</span>`,
				day.Year(), day.Month(), day.Day(), day.Year())
		} else {
			return kit.Format(`<span data-year="%d" data-month="%d" data-day="%d">%d月</span>`,
				day.Year(), day.Month(), day.Day(), day.Month())
		}
	}
	return kit.Format(`<span data-year="%d" data-month="%d" data-day="%d">%d</span>`,
		day.Year(), day.Month(), day.Day(), day.Day())
}

var Index = &ice.Context{Name: "team", Help: "团队中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"task": {Name: "task", Help: "任务", Value: kit.Data(kit.MDB_SHORT, "zone")},
		"plan": {Name: "plan", Help: "计划", Value: kit.Data(kit.MDB_SHORT, "zone",
			"head", []interface{}{"周日", "周一", "周二", "周三", "周四", "周五", "周六"},
			"template", kit.Dict(
				"day", `<div class="task {{.status}}" data-zone="%s" data-id="{{.id}}" data-begin_time="{{.begin_time}}">{{.status}} {{.name}}: {{.text}}</div>`,
				"week", `<div class="task {{.status}}" data-zone="%s" data-id="{{.id}}" data-begin_time="{{.begin_time}}" title="{{.text}}">{{.name}}</div>`,
				"month", `<div class="task {{.status}}" data-zone="%s" data-id="{{.id}}" data-begin_time="{{.begin_time}}" title="{{.text}}">{{.name}}</div>`,
				"year", `<div class="task {{.status}}" data-zone="%s" data-id="{{.id}}" data-begin_time="{{.begin_time}}">{{.name}}: {{.text}}</div>`,
			),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("task")
		}},

		"task": {Name: "task [zone [id [type [name [text args...]]]]]", Help: "任务", Meta: kit.Dict("remote", "you"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "zone", "action", "auto",
			kit.MDB_INPUT, "text", "name", "id", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "export":
					// 导出数据
					m.Option("cache.limit", -2)
					if f, p, e := kit.Create(arg[2]); m.Assert(e) {
						defer f.Close()

						w := csv.NewWriter(f)
						defer w.Flush()

						w.Write([]string{"begin_time", "close_time", "status", "type", "name", "text", "extra", "zone"})
						m.Richs(cmd, nil, kit.Select("*", arg, 3), func(key string, zone map[string]interface{}) {
							m.Grows(cmd, kit.Keys("hash", key), kit.Select("", arg, 4), kit.Select("", arg, 5), func(index int, task map[string]interface{}) {
								w.Write([]string{
									kit.Format(task["begin_time"]),
									kit.Format(task["close_time"]),
									kit.Format(task["status"]),
									kit.Format(task["type"]),
									kit.Format(task["name"]),
									kit.Format(task["text"]),
									kit.Format(task["extra"]),
									kit.Format(kit.Value(zone, "meta.zone")),
								})
							})
						})
						m.Log(ice.LOG_EXPORT, "%s", p)
					}

				case "import":
					// 导入数据
					m.CSV(m.Cmdx("nfs.cat", arg[2])).Table(func(index int, data map[string]string, head []string) {
						item := kit.Dict("time", data["time"],
							"type", data["type"], "name", data["name"], "text", data["text"], "extra", kit.UnMarshal(data["extra"]),
							"begin_time", data["begin_time"], "close_time", data["close_time"], "status", data["status"],
						)

						if m.Richs(cmd, nil, data["zone"], nil) == nil {
							// 添加领域
							m.Log(ice.LOG_CREATE, "zone: %s", data["zone"])
							m.Rich(cmd, nil, kit.Data("zone", data["zone"]))
						}

						m.Richs(cmd, nil, data["zone"], func(key string, value map[string]interface{}) {
							// 添加任务
							n := m.Grow(cmd, kit.Keys("hash", key), item)
							m.Log(ice.LOG_IMPORT, "%s: %d %s: %s", data["zone"], n, data["type"], data["name"])
						})
					})
				case "modify":
					// 任务修改
					m.Richs(cmd, nil, kit.Select(m.Option("zone"), arg, 6), func(key string, account map[string]interface{}) {
						m.Grows(cmd, kit.Keys("hash", key), "id", arg[5], func(index int, current map[string]interface{}) {
							m.Log(ice.LOG_MODIFY, "%s: %d %s: %s->%s", key, index, kit.Value(current, arg[2]), arg[2], arg[3])
							kit.Value(current, arg[2], arg[3])
						})
					})
				case "process":
					m.Richs(cmd, nil, kit.Select(m.Option("zone"), arg, 3), func(key string, account map[string]interface{}) {
						m.Grows(cmd, kit.Keys("hash", key), "id", kit.Select(m.Option("id"), arg, 2), func(index int, current map[string]interface{}) {
							if kit.Format(kit.Value(current, "status")) == "prepare" {
								m.Log(ice.LOG_MODIFY, "%s: %s %s: %s->%s", key, index, kit.Value(current, "status"), "status", "process")
								kit.Value(current, "begin_time", m.Time())
								kit.Value(current, "status", "process")
								m.Echo(arg[1])
							}
						})
					})
				case "finish", "cancel":
					m.Richs(cmd, nil, kit.Select(m.Option("zone"), arg, 3), func(key string, account map[string]interface{}) {
						m.Grows(cmd, kit.Keys("hash", key), "id", kit.Select(m.Option("id"), arg, 2), func(index int, current map[string]interface{}) {
							m.Log(ice.LOG_MODIFY, "%s: %d %s: %s->%s", key, index, kit.Value(current, "status"), "status", arg[1])
							kit.Value(current, "close_time", m.Time())
							kit.Value(current, "status", arg[1])
							m.Echo(arg[1])
						})
					})
				}
				return
			}

			if len(arg) == 0 {
				// 分类列表
				m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"], []string{"time", "count", "zone"})
				})
				return
			}

			if m.Richs(cmd, nil, arg[0], nil) == nil {
				// 添加分类
				m.Rich(cmd, nil, kit.Data("zone", arg[0]))
				m.Log(ice.LOG_CREATE, "zone: %s", arg[0])
			}

			field := []string{"begin_time", "id", "status", "type", "name", "text"}
			m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
				if len(arg) == 1 {
					// 任务列表
					m.Grows(cmd, kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						m.Push("", value, field)
					})
					m.Sort("time", "time_r")
					return
				}

				if len(arg) == 2 {
					// 任务详情
					m.Grows(cmd, kit.Keys("hash", key), "id", arg[1], func(index int, value map[string]interface{}) {
						m.Push("detail", value)
					})
					m.Sort("time", "time_r")
					return
				}

				if len(arg) < 5 {
					// 任务查询
					name, value := "type", arg[2]
					switch len(arg) {
					case 3:
						// 分类查询
						name, value = "type", arg[2]
					case 4:
						// 名称查询
						name, value = "name", arg[3]
					}
					m.Grows(cmd, kit.Keys("hash", key), name, value, func(index int, value map[string]interface{}) {
						m.Push("", value, field)
					})
					m.Sort("time", "time_r")
					return
				}

				// 词汇统计
				web.Count(m, cmd, "meta.word.type", arg[2])
				web.Count(m, cmd, "meta.word.name", arg[3])

				// 数据结构
				extra := kit.Dict()
				data := kit.Dict("type", arg[2], "name", arg[3], "text", arg[4], "extra", extra,
					"begin_time", m.Time(), "close_time", m.Time(), "status", "prepare",
				)

				// 扩展字段
				for i := 5; i < len(arg); i += 2 {
					switch arg[i] {
					case "begin_time", "close_time", "status":
						kit.Value(data, arg[i], arg[i+1])
					default:
						kit.Value(extra, arg[i], arg[i+1])
					}
				}

				// 添加任务
				n := m.Grow(cmd, kit.Keys("hash", key), data)
				m.Echo("%s: %d", kit.Value(value, "meta.zone"), n)
			})
		}},
		"plan": {Name: "plan day|week|month|year", Help: "计划", Meta: kit.Dict(
			"remote", "you", "display", "team/plan", "detail", []string{"process", "finish", "cancel"},
		), List: kit.List(
			kit.MDB_INPUT, "select", "name", "scale", "value", "day", "values", []string{"day", "week", "month", "months", "year", "long"}, "action", "auto",
			kit.MDB_INPUT, "text", "name", "begin_time", "figure", "date", "action", "auto",
			kit.MDB_INPUT, "text", "name", "end_time", "figure", "date", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cache.limit", -1)

			// 起始日期
			first := time.Now()
			if len(arg) > 1 {
				first = time.Unix(int64(kit.Time(arg[1])), 0)
			}
			first = DateZero(m, first)

			// 结束日期
			last := time.Now()
			if len(arg) > 2 {
				last = time.Unix(int64(kit.Time(arg[2])), 0)
			}
			last = DateZero(m, last)

			switch head := kit.Simple(m.Confv(cmd, "meta.head")); arg[0] {
			case "action":
				switch arg[1] {
				case "insert":
					// 创建任务
					m.Cmdy("task", arg[2], "", arg[3:])

				default:
					// 其它操作
					m.Cmdy("task", arg)
				}

			case "day":
				// 日计划
				for i := 6; i < 24; i++ {
					m.Push("hour", kit.Format("%02d", i))
					m.Push("task", "")
					m.Push("_id", "")
					m.Push("_zone", "")
				}

				match := first.Format("2006-01-02")
				template := m.Conf("plan", kit.Keys("meta.template", kit.Select("day", m.Option("template"))))
				m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
					m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						if now := kit.Format(value["begin_time"]); strings.Split(now, " ")[0] == match {
							b, _ := kit.Render(kit.Format(template, key), value)
							m.Push("hour", strings.Split(now, " ")[1][:2])
							m.Push("task", string(b))
							m.Push("_id", value["id"])
							m.Push("_zone", key)
						}
					})
				})
				m.Sort("hour", "int")

			case "week":
				// 周计划
				first = first.Add(-time.Duration((int64(first.Hour())*int64(time.Hour) + int64(first.Minute())*int64(time.Minute) + int64(first.Second())*int64(time.Second))))
				one := first.AddDate(0, 0, -int(first.Weekday()))
				end := first.AddDate(0, 0, 7-int(first.Weekday()))

				// 查询任务
				name := map[int][]string{}
				list := map[int][]map[string]interface{}{}
				m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
					m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						if t, e := time.ParseInLocation(ice.ICE_TIME, kit.Format(value["begin_time"]), time.Local); e == nil {
							if t.After(one) && t.Before(end) {
								index := t.Hour()*10 + int(t.Weekday())
								list[index] = append(list[index], value)
								name[index] = append(name[index], key)
							}
						}
					})
				})

				// 展示任务
				template := m.Conf("plan", kit.Keys("meta.template", kit.Select("week", m.Option("template"))))
				for i := 6; i < 24; i++ {
					m.Push("hour", kit.Format("%02d", i))
					for t := one; t.Before(end); t = t.AddDate(0, 0, 1) {
						index := i*10 + int(t.Weekday())

						note := []string{}
						for i, v := range list[index] {
							b, _ := kit.Render(kit.Format(template, name[index][i]), v)
							note = append(note, string(b))
						}
						m.Push(head[int(t.Weekday())], strings.Join(note, ""))
					}
				}

			case "month":
				fallthrough
			case "months":
				// 月计划
				one := first.AddDate(0, 0, -first.Day()+1)
				end := last.AddDate(0, 1, -last.Day()+1)
				template := m.Conf("plan", "meta.template.month")

				// 查询任务
				name := map[string][]string{}
				list := map[string][]map[string]interface{}{}
				m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
					m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						if t, e := time.ParseInLocation(ice.ICE_TIME, kit.Format(value["begin_time"]), time.Local); e == nil {
							if index := t.Format("2006-01-02"); t.After(one) && t.Before(end) {
								list[index] = append(list[index], value)
								name[index] = append(name[index], key)
							}
						}
					})
				})

				// 上月结尾
				last := one.AddDate(0, 0, -int(one.Weekday()))
				for day := last; day.Before(one); day = day.AddDate(0, 0, 1) {
					m.Push(head[int(day.Weekday())], ShowDay(m, day))
				}
				// 本月日期
				for day := one; day.Before(end); day = day.AddDate(0, 0, 1) {
					note := []string{ShowDay(m, day)}

					index := day.Format("2006-01-02")
					for i, v := range list[index] {
						b, _ := kit.Render(kit.Format(template, name[index][i]), v)
						note = append(note, string(b))
						// note = append(note, kit.Format(`%s: %s`, v["name"], v["text"]))
					}

					if len(note) > 1 {
						note[0] = kit.Format(`%s<sup class="more">%d</sup>`, note[0], len(note)-1)
					} else {
						note[0] = kit.Format(`%s<sup class="less">%s</sup>`, note[0], "")
					}

					if arg[0] == "month" {
						m.Push(head[int(day.Weekday())], strings.Join(note, ""))
					} else {
						m.Push(head[int(day.Weekday())], note[0])
					}
				}
				// 下月开头
				tail := end.AddDate(0, 0, 6-int(end.Weekday())+1)
				for day := end; end.Weekday() != 0 && day.Before(tail); day = day.AddDate(0, 0, 1) {
					m.Push(head[int(day.Weekday())], ShowDay(m, day))
				}

			case "year":
				// 年计划
				for i := 1; i < 13; i++ {
					m.Push("month", kit.Format("%02d", i))
					m.Push("task", "")
				}

				// 查询任务
				match := first.Format("2006")
				template := m.Conf("plan", kit.Keys("meta.template", kit.Select("year", m.Option("template"))))
				m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
					m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						if now := kit.Format(value["begin_time"]); now[0:4] == match && kit.Format(value["type"]) == "年度目标" {
							b, _ := kit.Render(kit.Format(template, key), value)
							m.Push("month", now[5:7])
							m.Push("task", string(b))
						}
					})
				})
				m.Sort("month", "int")

			case "long":
				// 长计划
				one := time.Unix(int64(kit.Time(kit.Select(kit.Format("%d-01-01", first.Year()-5), arg, 1))), 0)
				end := time.Unix(int64(kit.Time(kit.Select(kit.Format("%d-12-31", first.Year()+5), arg, 2))), 0)
				for day := one; day.Before(end); day = day.AddDate(1, 0, 0) {
					m.Push("year", day.Year())
					m.Push("task", "")
				}

				// 查询任务
				template := m.Conf("plan", kit.Keys("meta.template", kit.Select("year", m.Option("template"))))
				m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
					m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						if t, e := time.ParseInLocation(ice.ICE_TIME, kit.Format(value["begin_time"]), time.Local); e == nil {
							if t.After(one) && t.Before(end) && kit.Format(value["type"]) == "年度目标" {
								b, _ := kit.Render(kit.Format(template, key), value)
								m.Push("year", t.Year())
								m.Push("task", string(b))
							}
						}
					})
				})
				m.Sort("year", "int")
			}
		}},
		"stat": {Name: "stat", Help: "统计", Meta: kit.Dict("remote", "you"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "begin_time", "figure", "date", "action", "auto",
			kit.MDB_INPUT, "text", "name", "end_time", "figure", "date", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			begin_time := kit.Time(kit.Select("1990-07-30", arg, 0))
			end_time := kit.Time(kit.Select("1990-07-30", arg, 1))
			now_time := kit.Time(m.Time())

			m.Option("cache.limit", -1)
			m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
				stat := map[string]int{}
				m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
					if len(arg) > 1 && kit.Time(kit.Format(value["begin_time"])) > end_time {
						return
					}
					if len(arg) > 0 && kit.Time(kit.Format(value["begin_time"])) < begin_time {
						return
					}

					if stat[kit.Format(value["status"])] += 1; value["status"] != "prepare" {
						use := kit.Time(kit.Format(value["close_time"])) - kit.Time(kit.Format(value["begin_time"]))
						if value["status"] == "process" {
							use = now_time - kit.Time(kit.Format(value["begin_time"]))
						}

						stat["sum"] += use
						if use > stat["max"] {
							stat["max"] = use
						}
						if use < stat["min"] {
							stat["min"] = use
						}
					}
					stat["total"] += 1
				})

				m.Push("zone", kit.Value(value, "meta.zone"))
				for _, k := range []string{"prepare", "process", "finish", "total"} {
					m.Push(k, stat[k])
				}
				m.Push("sum", kit.FmtTime(int64(stat["sum"])*int64(time.Second)))
				if stat["finish"] == 0 {
					stat["finish"] = 1
				}
				m.Push("avg", kit.FmtTime(int64(stat["sum"]/stat["finish"])*int64(time.Second)))
				m.Push("min", kit.FmtTime(int64(stat["min"])*int64(time.Second)))
				m.Push("max", kit.FmtTime(int64(stat["max"])*int64(time.Second)))
			})
		}},
		"miss": {Name: "miss zone type name text", Help: "任务", Meta: kit.Dict("remote", "you"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "zone", "figure", "key", "action", "auto",
			kit.MDB_INPUT, "text", "name", "type", "figure", "key",
			kit.MDB_INPUT, "text", "name", "name", "figure", "key",
			kit.MDB_INPUT, "button", "name", "添加",
			kit.MDB_INPUT, "textarea", "name", "text",
			kit.MDB_INPUT, "text", "name", "location", "figure", "key", "cb", "location", "className", "opts",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "input":
					// 输入补全
					switch arg[2] {
					case "zone":
						m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
							m.Push("zone", kit.Value(value, "meta.zone"))
							m.Push("count", kit.Value(value, "meta.count"))
						})
					case "type", "name":
						m.Confm("task", kit.Keys("meta.word", arg[2]), func(key string, value string) {
							m.Push(arg[2], key)
							m.Push("count", value)
						})
					}
					m.Sort("count", "int_r")
					return
				}
			}

			if len(arg) < 2 {
				// 查询任务
				m.Cmdy("task", arg)
				return
			}
			// 创建任务
			m.Cmdy("task", arg[0], "", arg[1:])
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
