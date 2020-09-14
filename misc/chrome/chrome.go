package crx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const CHROME = "chrome"

var Index = &ice.Context{Name: "chrome", Help: "浏览器",
	Configs: map[string]*ice.Config{
		CHROME: {Name: "chrome", Help: "浏览器", Value: kit.Data(
			kit.MDB_SHORT, "name", "history", "url.history",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		CHROME: {Name: "chrome wid=auto url auto 编译:button 下载:button", Help: "浏览器", Action: map[string]*ice.Action{
			"compile": {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(web.SPACE, CHROME, CHROME, arg)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
