package crx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"
)

const CHROME = "chrome"

var Index = &ice.Context{Name: "chrome", Help: "浏览器",
	Configs: map[string]*ice.Config{
		CHROME: {Name: "chrome", Help: "浏览器", Value: kit.Data(
			kit.MDB_SHORT, "name", web.FAVOR, "url.history",
		)},
	},
	Commands: map[string]*ice.Command{
		"/crx": {Name: "/crx", Help: "/crx", Action: map[string]*ice.Action{
			"history": {Name: "history", Help: "历史记录", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPIDE, "dev", "/code/chrome/favor", "cmds", "insert", "name", arg[1], "note", arg[2],
					"tab", m.Conf(CHROME, "meta.favor"), "sid", m.Option("sid"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"/favor": {Name: "/favor", Help: "收藏", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.FAVOR, mdb.INSERT, m.Option("tab"), web.SPIDE, m.Option("name"), m.Option("note"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		CHROME: {Name: "chrome", Help: "浏览器", List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "action", "auto",
			kit.MDB_INPUT, "text", "name", "wid", "action", "auto",
			kit.MDB_INPUT, "text", "name", "url",
			kit.MDB_INPUT, "button", "name", "查看",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Meta: kit.Dict("detail", []string{"编辑", "goBack", "goForward", "duplicate", "reload", "remove"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 窗口列表
				m.Richs(web.SPACE, nil, "*", func(key string, value map[string]interface{}) {
					if kit.Format(value["type"]) == "chrome" {
						m.Push(key, value, []string{"time", "name"})
					}
				})
				return
			}
			if arg[0] == "action" {
				// 命令转换
				m.Cmdy(web.SPACE, m.Option("name"), "tabs", m.Option("tid"), arg[1])
				arg = []string{m.Option("name"), m.Option("wid")}
			}
			// 下发命令
			m.Cmdy(web.SPACE, arg[0], "wins", arg[1:])
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
				m.Cmdy(web.SPACE, m.Option("name"), "cookie", arg[1:])
				arg = []string{m.Option("name"), m.Option("id")}
			}
			// 下发命令
			m.Cmdy(web.SPACE, arg[0], "cookie", arg[1:])
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
				m.Cmdy(web.SPACE, m.Option("name"), "bookmark", arg[1:])
				arg = []string{m.Option("name"), m.Option("id")}
			}
			// 下发命令
			m.Cmdy(web.SPACE, arg[0], "bookmark", arg[1:])
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
