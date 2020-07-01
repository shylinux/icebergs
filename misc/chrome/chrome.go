package crx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"
)

const CHROME = "chrome"
const HISTORY = "history"
const BOOKMARK = "bookmark"

var Index = &ice.Context{Name: "chrome", Help: "浏览器",
	Configs: map[string]*ice.Config{
		CHROME: {Name: "chrome", Help: "浏览器", Value: kit.Data(
			kit.MDB_SHORT, "name", web.FAVOR, "url.history",
		)},
	},
	Commands: map[string]*ice.Command{
		"/crx": {Name: "/crx", Help: "/crx", Action: map[string]*ice.Action{
			web.HISTORY: {Name: "history", Help: "历史记录", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPIDE, "dev", "/code/chrome/favor", "cmds", mdb.INSERT,
					"tab", m.Conf(CHROME, "meta.favor"), "name", arg[1], "note", arg[2],
					"sid", m.Option("sid"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/favor": {Name: "/favor", Help: "收藏", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.FAVOR, mdb.INSERT, m.Option("tab"), web.SPIDE, m.Option("name"), m.Option("note"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		CHROME: {Name: "chrome name=chrome wid=auto url auto", Help: "浏览器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 窗口列表
				m.Richs(web.SPACE, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					if kit.Format(value[kit.MDB_TYPE]) == CHROME {
						m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_NAME})
					}
				})
				return
			}
			// 下发命令
			m.Cmdy(web.SPACE, arg[0], CHROME, arg[1:])
		}},
		BOOKMARK: {Name: "bookmark name=chrome auto", Help: "书签", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 窗口列表
				m.Cmdy(CHROME)
				return
			}
			// 下发命令
			m.Cmdy(web.SPACE, arg[0], BOOKMARK, arg[1:])
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
