package crx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "chrome", Help: "浏览器",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"chrome": {Name: "chrome", Help: "chrome", Value: kit.Data(kit.MDB_SHORT, "name", "history", "url.history")},
	},
	Commands: map[string]*ice.Command{

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

		ice.WEB_LOGIN: {Name: "_login", Help: "_login", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("you", "")
			m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
				// 查找空间
				m.Option("you", value["you"])
			})

			m.Info("%s %s cmd: %v sub: %v", m.Option("you"), m.Option(ice.MSG_USERURL), m.Optionv("cmds"), m.Optionv("sub"))
		}},
		"/help": {Name: "/help", Help: "帮助", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("help")
		}},
		"/login": {Name: "/login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("login", "init", c.Name)
		}},
		"/logout": {Name: "/logout", Help: "登出", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("login", "exit")
		}},

		"/favor": {Name: "/favor", Help: "收藏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				// 添加收藏
				cmds := []string{ice.WEB_FAVOR, m.Option("tab"), ice.TYPE_SPIDE, m.Option("note"), arg[0]}
				if m.Cmdy(cmds); m.Option("you") != "" {
					m.Cmdy(ice.WEB_SPACE, m.Option("you"), cmds)
				}
				return
			}
		}},

		"/crx": {Name: "/crx", Help: "/crx", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Info("what %v", m.Option("sid"))
			switch arg[0] {
			case "login":
				m.Cmdy(ice.WEB_SPIDE, "dev", "msg", "/code/chrome/login", "sid", m.Option("sid"))

			case "bookmark":
				m.Cmdy(ice.WEB_SPIDE, "dev", "/code/chrome/favor", "cmds", arg[2], "note", arg[3],
					"tab", kit.Select(m.Conf("chrome", "meta.history"), arg, 4), "sid", m.Option("sid"), "type", "spide")

			case "history":
				m.Cmdy(ice.WEB_SPIDE, "dev", "/code/chrome/favor", "cmds", arg[2], "note", arg[3],
					"tab", m.Conf("chrome", "meta.history"), "sid", m.Option("sid"))
			}
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
