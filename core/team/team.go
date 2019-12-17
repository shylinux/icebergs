package team

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "team", Help: "团队模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		"task": {Name: "task", Help: "任务", List: []interface{}{
			map[string]interface{}{"type": "select", "value": "create", "values": "create action cancel finish"},
			map[string]interface{}{"type": "text", "value": "", "name": "name"},
			map[string]interface{}{"type": "text", "value": "", "name": "text"},
			map[string]interface{}{"type": "button", "value": "创建"},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			switch arg[0] {
			case "create":
				m.Log("waht", "%v", m.Conf("web.chat.group", "hash."+m.Option("sess.river")+".task"))
				meta := m.Grow("web.chat.group", []string{"hash", m.Option("sess.river"), "task"}, map[string]interface{}{
					"name":       arg[1],
					"text":       kit.Select("", arg, 2),
					"status":     "prepare",
					"begin_time": m.Time(),
					"close_time": m.Time("3h"),
				})
				m.Log("info", "create task %v", kit.Format(meta))
				m.Log("waht", "%v %v", meta["count"], m.Conf("web.chat.group", "hash."+m.Option("sess.river")+".task"))
				m.Echo("%v", meta["count"])
			case "action":
			case "cancel":
			}
		}},
		"process": {Name: "process", Help: "任务进度", Meta: map[string]interface{}{
			"detail": []string{"准备", "开始", "取消", "完成"},
		}, List: []interface{}{
			map[string]interface{}{"type": "text", "value": "0", "name": "offend"},
			map[string]interface{}{"type": "text", "value": "10", "name": "limit"},
			map[string]interface{}{"type": "text", "value": "", "name": "match"},
			map[string]interface{}{"type": "text", "value": "", "name": "value"},
			map[string]interface{}{"type": "button", "value": "查看"},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			prefix := []string{"mdb.update", "web.chat.group", "hash." + m.Option("sess.river") + ".task", arg[0]}
			switch arg[1] {
			case "准备":
				m.Cmd(prefix, "status", arg[1],
					"extra.list.-2.time", m.Time(),
					"extra.list.-3.status", arg[1],
				)
				arg = arg[4:]
			case "开始":
				m.Cmd(prefix, "status", arg[1], "begin_time", m.Time(),
					"extra.list.-2.time", m.Time(),
					"extra.list.-3.status", arg[1],
				)
				arg = arg[4:]
			case "取消":
				m.Cmd(prefix, "status", arg[1], "close_time", m.Time(),
					"extra.list.-2.time", m.Time(),
					"extra.list.-3.status", arg[1],
				)
				arg = arg[4:]
			case "完成":
				m.Cmd(prefix, "status", arg[1], "close_time", m.Time(),
					"extra.list.-2.time", m.Time(),
					"extra.list.-3.status", arg[1],
				)
				arg = arg[4:]
			}

			m.Option("cache.offend", kit.Select("0", arg, 0))
			m.Option("cache.limit", kit.Select("10", arg, 1))
			m.Option("cache.match", kit.Select("", arg, 2))
			m.Option("cache.value", kit.Select("", arg, 3))
			m.Grows("web.chat.group", []string{"hash", m.Option("sess.river"), "task"}, func(index int, value map[string]interface{}) {
				m.Push("id", value, []string{"id", "status", "begin_time", "close_time", "name", "text"})
			})
		}},
	},
}

func init() { web.Index.Register(Index, &web.WEB{}) }
