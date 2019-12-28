package team

import (
	"fmt"
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
	"time"
)

var Index = &ice.Context{Name: "team", Help: "团队模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.APP_MISS: {Name: "miss", Help: "任务", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", "team.json")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "team.json", "web.team.miss")
		}},

		ice.APP_MISS: {Name: "miss", Help: "任务", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) > 1 {
				switch arg[1] {
				case "modify":
					// 修改任务
					m.Grows(ice.APP_MISS, nil, "id", arg[0], func(index int, value map[string]interface{}) {
						value[arg[2]] = arg[3]
					})
					arg = arg[:0]
				}
			}

			if len(arg) == 0 {
				// 任务列表
				m.Grows(ice.APP_MISS, nil, "", "", func(index int, value map[string]interface{}) {
					m.Push(kit.Format(index), value, []string{"begin_time", "close_time", "status", "id", "type", "name", "text"})
				})
				return
			}

			// 添加任务
			h := m.Grow(ice.APP_MISS, nil, kit.Dict(
				kit.MDB_NAME, arg[0],
				kit.MDB_TYPE, kit.Select("开发", arg, 1),
				kit.MDB_TEXT, kit.Select("功能开发", arg, 2),
				"status", kit.Select("准备中", arg, 3),
				"begin_time", m.Time(), "close_time", m.Time(),
			))
			m.Info("miss: %d", h)
			m.Echo("%d", h)
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
			"display", "github.com/shylinux/icebergs/core/team/miss",
			"detail", []string{"回退", "前进", "取消", "完成"},
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) > 0 {
				m.Grows(ice.APP_MISS, nil, "id", arg[0], func(index int, value map[string]interface{}) {
					switch value["status"] {
					case "准备中":
						switch arg[1] {
						case "开始":
							value["status"] = "进行中"
							value["begin_time"] = m.Time()
							value["close_time"] = m.Time("30m")
						case "取消":
							value["status"] = "已取消"
							value["close_time"] = m.Time()
						case "完成":
							value["status"] = "已完成"
							value["close_time"] = m.Time()
						}
					case "进行中":
						switch arg[1] {
						case "准备":
							value["status"] = "准备中"
							value["begin_time"] = m.Time()
							value["close_time"] = m.Time()
						case "取消":
							value["status"] = "已取消"
							value["close_time"] = m.Time()
						case "完成":
							value["status"] = "已完成"
							value["close_time"] = m.Time()
						}
					}
				})
			}

			m.Push("准备中", "")
			m.Push("开发中", "")
			m.Push("测试中", "")
			m.Push("发布中", "")
			m.Push("已取消", "")
			m.Push("已完成", "")
			m.Grows(ice.APP_MISS, nil, "", "", func(index int, value map[string]interface{}) {
				m.Push(kit.Format(value["status"]),
					kit.Format(`<span title="%v" data-id="%v">%v</span>`,
						kit.Format("%s-%s\n%s", value["begin_time"], value["close_time"], value["text"]),
						value["id"], value["name"]))
			})
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
