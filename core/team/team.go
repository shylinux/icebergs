package team

import (
	"fmt"
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
	"time"
)

var Index = &ice.Context{Name: "team", Help: "团队中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.APP_MISS: {Name: "miss", Help: "任务", Value: kit.Data(
			"mis", []interface{}{"已取消", "准备中", "开发中", "测试中", "发布中", "已完成"}, "fsm", kit.Dict(
				"准备中", kit.Dict("next", "开发中"),
				"开发中", kit.Dict("next", "测试中", "prev", "准备中"),
				"测试中", kit.Dict("next", "发布中", "prev", "开发中"),
				"发布中", kit.Dict("next", "已完成", "prev", "测试中"),
				"已完成", kit.Dict(),
				"已取消", kit.Dict(),
			),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Watch(ice.MISS_CREATE, ice.APP_MISS)
			m.Cmd(ice.CTX_CONFIG, "load", "team.json")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "team.json", "web.team.miss")
		}},

		ice.APP_MISS: {Name: "miss", Help: "任务", Meta: kit.Dict(
			"remote", "you",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			hot := kit.Select(ice.FAVOR_MISS, m.Option("hot"))
			if len(arg) > 1 {
				switch arg[1] {
				case "modify":
					// 修改任务
					m.Richs(ice.WEB_FAVOR, nil, hot, func(key string, value map[string]interface{}) {
						m.Grows(ice.WEB_FAVOR, kit.Keys("hash", key), "id", arg[0], func(index int, value map[string]interface{}) {
							m.Log(ice.LOG_MODIFY, "%s: %s->%s", arg[2], arg[4], arg[3])
							kit.Value(value, arg[2], arg[3])
						})
					})
					arg = arg[:0]
				}
			}

			if len(arg) == 0 {
				// 任务列表
				m.Richs(ice.WEB_FAVOR, nil, hot, func(key string, value map[string]interface{}) {
					m.Grows(ice.WEB_FAVOR, kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						m.Push(kit.Format(index), value, []string{"extra.begin_time", "extra.close_time", "extra.status", "id", "type", "name", "text"})
					})
				})
				return
			}

			// 添加任务
			m.Cmdy(ice.WEB_FAVOR, hot, ice.TYPE_DRIVE, arg[0], arg[1],
				"begin_time", m.Time(), "close_time", m.Time(),
				"status", kit.Select("准备中", arg, 3),
			)
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
		"stat": {Name: "stat", Help: "统计", Meta: kit.Dict(
		// "display", "github.com/shylinux/icebergs/core/team/stat",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Push("weekly", 10)
			m.Push("month", 100)
			m.Push("year", 1000)
		}},
		"progress": {Name: "progress", Help: "进度", Meta: kit.Dict(
			"remote", "you",
			"display", "github.com/shylinux/icebergs/core/team/miss",
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
							}

						case "回退":
							if prev := m.Conf(ice.APP_MISS, kit.Keys("meta.fsm", value["status"], "prev")); prev != "" {
								value["status"] = prev
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
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
