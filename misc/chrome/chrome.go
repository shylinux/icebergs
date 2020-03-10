package crx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "chrome", Help: "浏览器",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"chrome": {Name: "chrome", Help: "chrome", Value: kit.Data(kit.MDB_SHORT, "name", "history", "url.history")},
	},
	Commands: map[string]*ice.Command{
		"/crx": {Name: "/crx", Help: "/crx", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "history":
				m.Cmdy(ice.WEB_FAVOR, m.Conf("chrome", "meta.history"), "spide", arg[3], arg[2])
			}
		}},
		"chrome": {Name: "chrome", Help: "浏览器", List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "action", "auto",
			kit.MDB_INPUT, "text", "name", "wid", "action", "auto",
			kit.MDB_INPUT, "text", "name", "url",
			kit.MDB_INPUT, "button", "name", "查看",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Meta: kit.Dict("detail", []string{"编辑", "goBack", "goForward", "duplicate", "reload", "remove"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 窗口列表
				m.Richs(ice.WEB_SPACE, nil, "*", func(key string, value map[string]interface{}) {
					if kit.Format(value["type"]) == "chrome" {
						m.Push(key, value, []string{"time", "name"})
					}
				})
				return
			}
			if arg[0] == "action" {
				// 命令转换
				m.Cmdy(ice.WEB_SPACE, m.Option("name"), "tabs", m.Option("tid"), arg[1])
				arg = []string{m.Option("name"), m.Option("wid")}
			}
			// 下发命令
			m.Cmdy(ice.WEB_SPACE, arg[0], "wins", arg[1:])
		}},
		"cookie": {Name: "cookie", Help: "数据", List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "action", "auto",
			kit.MDB_INPUT, "text", "name", "id", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Meta: kit.Dict("detail", []string{"编辑", "删除"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 窗口列表
				m.Cmdy("chrome")
				return
			}
			if arg[0] == "action" {
				// 命令转换
				m.Cmdy(ice.WEB_SPACE, m.Option("name"), "cookie", arg[1:])
				arg = []string{m.Option("name"), m.Option("id")}
			}
			// 下发命令
			m.Cmdy(ice.WEB_SPACE, arg[0], "cookie", arg[1:])
		}},
		"bookmark": {Name: "bookmark", Help: "书签", List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "action", "auto",
			kit.MDB_INPUT, "text", "name", "id", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Meta: kit.Dict("detail", []string{"编辑", "删除"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 窗口列表
				m.Cmdy("chrome")
				return
			}
			if arg[0] == "action" {
				// 命令转换
				m.Cmdy(ice.WEB_SPACE, m.Option("name"), "bookmark", arg[1:])
				arg = []string{m.Option("name"), m.Option("id")}
			}
			// 下发命令
			m.Cmdy(ice.WEB_SPACE, arg[0], "bookmark", arg[1:])
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
