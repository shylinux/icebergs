package team

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"fmt"
	"strings"
	"time"
)

var Index = &ice.Context{Name: "team", Help: "团队中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"task":     {Name: "task", Help: "任务", Value: kit.Data(kit.MDB_SHORT, "zone")},
		"location": {Name: "location", Help: "位置", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", kit.Keys(m.Cap(ice.CTX_FOLLOW), "json"))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", kit.Keys(m.Cap(ice.CTX_FOLLOW), "json"), kit.Keys(m.Cap(ice.CTX_FOLLOW), "task"))
		}},

		"task": {Name: "task", Help: "任务", Meta: kit.Dict("remote", "you"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "zone", "action", "auto",
			kit.MDB_INPUT, "text", "name", "id", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cache.limit", "10000")

			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "modify":
					m.Richs(cmd, nil, m.Option("zone"), func(key string, account map[string]interface{}) {
						m.Grows(cmd, kit.Keys("hash", key), "id", arg[5], func(index int, current map[string]interface{}) {
							kit.Value(current, arg[2], arg[3])
						})
					})
				}
				return
			}

			if len(arg) == 0 {
				// 任务列表
				m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"])
				})
				return
			}

			if m.Richs(cmd, nil, arg[0], nil) == nil {
				// 添加任务
				m.Rich(cmd, nil, kit.Data("zone", arg[0]))
				m.Log(ice.LOG_CREATE, "zone: %s", arg[0])
			}

			m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
				field := []string{"begin_time", "id", "status", "type", "name", "text"}
				if len(arg) == 1 {
					// 任务列表
					m.Grows(cmd, kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						m.Push("", value, field)
					})
					m.Sort("time", "time_r")
					return
				}
				if len(arg) == 2 {
					// 消费详情
					m.Grows(cmd, kit.Keys("hash", key), "id", arg[1], func(index int, value map[string]interface{}) {
						m.Push("detail", value)
					})
					m.Sort("time", "time_r")
					return
				}
				if len(arg) < 5 {
					name, value := "type", arg[2]
					switch len(arg) {
					case 3:
						// 消费分类
						name, value = "type", arg[2]
					case 4:
						// 消费对象
						name, value = "name", arg[3]
					}
					m.Grows(cmd, kit.Keys("hash", key), name, value, func(index int, value map[string]interface{}) {
						m.Push("", value, []string{"time", "id", "status", "type", "name", "text"})
					})
					m.Sort("time", "time_r")
					return
				}

				// 添加任务
				extra := kit.Dict()
				data := kit.Dict("type", arg[2], "name", arg[3], "text", arg[4],
					"begin_time", m.Time(), "close_time", m.Time(),
					"status", "prepare", "extra", extra,
				)

				count := kit.Int(m.Conf("task", kit.Keys("meta.word", "type", arg[2])))
				m.Conf("task", kit.Keys("meta.word", "type", arg[2]), count+1)
				count = kit.Int(m.Conf("task", kit.Keys("meta.word", "name", arg[3])))
				m.Conf("task", kit.Keys("meta.word", "name", arg[3]), count+1)

				for i := 5; i < len(arg); i += 2 {
					switch arg[i] {
					case "begin_time", "close_time", "status":
						kit.Value(data, arg[i], arg[i+1])
					default:
						kit.Value(extra, arg[i], arg[i+1])
					}
				}
				n := m.Grow(cmd, kit.Keys("hash", key), data)
				m.Echo("%s: %d", key, n)
			})
		}},
		"plan": {Name: "plan day|week|month|year", Help: "计划", Meta: kit.Dict("display", "team/plan"), List: kit.List(
			kit.MDB_INPUT, "select", "name", "scale", "values", []string{"day", "week", "month"}, "action", "auto",
			kit.MDB_INPUT, "text", "name", "begin_time", "action", "auto", "figure", "date",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "day")
			}

			// 起始日期
			first := time.Now()
			if len(arg) > 1 {
				first = time.Unix(int64(kit.Time(arg[1])), 0)
			}

			// 结束日期
			last := time.Now()
			if len(arg) > 2 {
				last = time.Unix(int64(kit.Time(arg[2])), 0)
			}

			meta := map[int]string{0: "周日", 1: "周一", 2: "周二", 3: "周三", 4: "周四", 5: "周五", 6: "周六"}

			switch arg[0] {
			case "action":
				switch arg[1] {
				case "insert":
					m.Cmdy("task", arg[2], "", arg[3:])

				case "modify":
					switch arg[2] {
					case "begin_time":
						m.Richs("task", nil, arg[6], func(key string, value map[string]interface{}) {
							m.Grows("task", kit.Keys("hash", key), "id", arg[5], func(index int, value map[string]interface{}) {
								m.Log(ice.LOG_MODIFY, "%s: %s begin_time: %s", arg[6], arg[5], arg[3])
								value["begin_time"] = arg[3]
							})
						})
					}
				}

			case "day":
				for i := 6; i < 24; i++ {
					m.Push("hour", i)
					m.Push("task", "")
				}

				match := first.Format("2006-01-02")
				m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
					m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						if now := kit.Format(value["begin_time"]); strings.Split(now, " ")[0] == match {
							m.Push("hour", strings.Split(now, " ")[1][:2])
							m.Push("task", kit.Format(`<div class="task" data-name="%s" data-id="%d" data-begin_time="%s">%s: %s</div>`,
								key, kit.Int(value["id"]), value["begin_time"], value["name"], value["text"]))
						}
					})
				})
				m.Sort("hour", "int")

			case "week":
				one := first.AddDate(0, 0, -int(first.Weekday()))
				end := first.AddDate(0, 0, 7-int(first.Weekday()))

				list := map[int][]map[string]interface{}{}
				name := map[int][]string{}
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

				for i := 6; i < 24; i++ {
					m.Push("hour", i)
					for t := one; t.Before(end); t = t.AddDate(0, 0, 1) {
						index := i*10 + int(t.Weekday())
						note := []string{}
						for i, v := range list[index] {
							note = append(note, kit.Format(`<div class="task" data-name="%s" data-id="%d" data-begin_time="%s" title="%s">%s</div>`,
								name[index][i], kit.Int(v["id"]), v["begin_time"], v["text"], v["name"]))
						}
						m.Push(meta[int(t.Weekday())], strings.Join(note, ""))
					}
				}

			case "month":
				// 本月日期
				one := first.AddDate(0, 0, -first.Day()+1)
				end := last.AddDate(0, 1, -last.Day()+1)

				list := map[string][]map[string]interface{}{}
				m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
					m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						if t, e := time.ParseInLocation(ice.ICE_TIME, kit.Format(value["begin_time"]), time.Local); e == nil {
							if t.After(one) && t.Before(end) {
								index := t.Format("2006-01-02")
								list[index] = append(list[index], value)
							}
						}
					})
				})

				// 上月结尾
				head := one.AddDate(0, 0, -int(one.Weekday()))
				for day := head; day.Before(one); day = day.AddDate(0, 0, 1) {
					m.Push(meta[int(day.Weekday())], day.Day())
				}
				// 本月日期
				for day := one; day.Before(end); day = day.AddDate(0, 0, 1) {
					note := []string{}
					if day.Day() == 1 {
						note = append(note, kit.Format("%d月", day.Month()))
					} else {
						note = append(note, kit.Format("%d", day.Day()))
					}

					index := day.Format("2006-01-02")
					for _, v := range list[index] {
						note = append(note, kit.Format(`%s: %s`, v["name"], v["text"]))
					}
					if len(note) > 1 {
						note[0] = kit.Format(`<div title="%s">%s<sup class="more">%d<sup><div>`, strings.Join(note[1:], "\n"), note[0], len(note)-1)
					} else {
						note[0] = kit.Format(`%s<sup class="less">%s<sup>`, note[0], "")
					}
					m.Push(meta[int(day.Weekday())], note[0])

				}
				// 下月开头
				tail := end.AddDate(0, 0, 6-int(end.Weekday())+1)
				for day := end; end.Weekday() != 0 && day.Before(tail); day = day.AddDate(0, 0, 1) {
					m.Push(meta[int(day.Weekday())], day.Day())
				}

			case "year":
			}
		}},
		"stat": {Name: "stat", Help: "统计", Meta: kit.Dict(), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Richs("task", nil, kit.Select("*", arg, 0), func(key string, value map[string]interface{}) {
				stat := map[string]int{}
				m.Grows("task", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
					stat[kit.Format(value["status"])] += 1
				})
				m.Push("task", kit.Value(value, "meta.task"))
				for _, k := range []string{"prepare", "process", "cancel", "finish"} {
					m.Push(k, stat[k])
				}
			})
		}},
		"date": {Name: "date", Help: "日历", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			show := map[int]string{0: "周日", 1: "周一", 2: "周二", 3: "周三", 4: "周四", 5: "周五", 6: "周六"}

			space := m.Options("space")
			today := time.Now()
			now := today
			n := kit.Int(kit.Select("1", m.Option("count")))

			cur := now
			for i := 0; i < n; i, now = i+1, now.AddDate(0, 1, 0) {
				begin := time.Unix(now.Unix()-int64(now.Day()-1)*24*3600, 0)
				last := time.Unix(begin.Unix()-int64(begin.Weekday())*24*3600, 0)
				cur = last

				if last.Month() != now.Month() {
					for month := cur.Month(); cur.Month() == month; cur = cur.AddDate(0, 0, 1) {
						if space || i == 0 {
							m.Push(show[int(cur.Weekday())], "")
						}
					}
				}
				for month := cur.Month(); cur.Month() == month; cur = cur.AddDate(0, 0, 1) {
					data := fmt.Sprintf("%d", cur.Day())
					if cur.Year() == today.Year() && cur.YearDay() == today.YearDay() {
						data = fmt.Sprintf(">%d<", cur.Day())
					}
					if cur.Day() == 1 {
						if cur.Month() == 1 {
							data = fmt.Sprintf("%d年", cur.Year())
						} else {
							data = fmt.Sprintf("%d月", cur.Month())
						}
					}
					m.Push(show[int(cur.Weekday())], data)
				}
				if space || i == n-1 {
					for ; cur.Weekday() > 0; cur = cur.AddDate(0, 0, 1) {
						m.Push(show[int(cur.Weekday())], "")
					}
				}
			}
		}},
		"progress": {Name: "progress", Help: "进度", Meta: kit.Dict(
			"remote", "you", "display", "team/miss",
			"detail", []string{"回退", "前进", "取消", "完成"},
		), List: kit.List(
			kit.MDB_INPUT, "text", "value", "",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			hot := kit.Select(ice.FAVOR_MISS, m.Option("hot"))
			if len(arg) > 0 {
				m.Richs(ice.WEB_FAVOR, nil, hot, func(key string, value map[string]interface{}) {
					m.Grows(ice.WEB_FAVOR, kit.Keys("hash", key), "id", arg[0], func(index int, value map[string]interface{}) {
						switch value = value["extra"].(map[string]interface{}); arg[1] {
						case "前进":
							if value["status"] == "准备中" {
								value["begin_time"] = m.Time()
								value["close_time"] = m.Time("30m")
							}
							if next := m.Conf(ice.APP_MISS, kit.Keys("meta.fsm", value["status"], "next")); next != "" {
								value["status"] = next
								kit.Value(value, "change.-2", kit.Dict("time", m.Time(), "status", next))
							}

						case "回退":
							if prev := m.Conf(ice.APP_MISS, kit.Keys("meta.fsm", value["status"], "prev")); prev != "" {
								value["status"] = prev
								kit.Value(value, "change.-2", kit.Dict("time", m.Time(), "status", prev))
							}

						case "取消":
							value["status"] = "已取消"
							value["close_time"] = m.Time()

						case "完成":
							value["status"] = "已完成"
							value["close_time"] = m.Time()
						}
					})
				})
			}

			m.Confm(ice.APP_MISS, "meta.mis", func(index int, value string) {
				m.Push(value, "")
			})
			m.Richs(ice.WEB_FAVOR, nil, hot, func(key string, value map[string]interface{}) {
				m.Grows(ice.WEB_FAVOR, kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
					m.Push(kit.Format(kit.Value(value, "extra.status")),
						kit.Format(`<span title="%v" data-id="%v">%v</span>`,
							kit.Format("%s-%s\n%s", kit.Value(value, "extra.begin_time"), kit.Value(value, "extra.close_time"), value["text"]),
							value["id"], value["name"]))
				})
			})
		}},

		"miss": {Name: "miss zone type name text", Help: "任务", List: kit.List(
			kit.MDB_INPUT, "text", "name", "zone", "action", "auto", "figure", "key",
			kit.MDB_INPUT, "text", "name", "type", "figure", "key",
			kit.MDB_INPUT, "text", "name", "name", "figure", "key",
			kit.MDB_INPUT, "button", "name", "添加",
			kit.MDB_INPUT, "textarea", "name", "text",
			kit.MDB_INPUT, "text", "name", "location", "figure", "key", "cb", "location",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "input":
					switch arg[2] {
					case "type", "name":
						m.Confm("task", kit.Keys("meta.word", arg[2]), func(key string, value string) {
							m.Push(arg[2], key)
							m.Push("count", value)
						})
						m.Sort("count", "int_r")
					case "zone":
						m.Richs("task", nil, "*", func(key string, value map[string]interface{}) {
							m.Push("zone", kit.Value(value, "meta.zone"))
							m.Push("count", kit.Value(value, "meta.count"))
						})
					}
					return
				}
			}

			if len(arg) < 2 {
				m.Cmdy("task", arg)
				return
			}
			m.Cmd("task", arg[0], "", arg[1:])
			m.Cmdy("task", arg[0])
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
