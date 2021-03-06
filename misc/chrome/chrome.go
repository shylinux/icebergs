package crx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const CHROME = "chrome"

var Index = &ice.Context{Name: CHROME, Help: "浏览器",
	Configs: map[string]*ice.Config{
		CHROME: {Name: CHROME, Help: "浏览器", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		CHROME: {Name: "chrome wid tid url auto 启动 构建 下载", Help: "浏览器", Action: map[string]*ice.Action{
			"install": {Name: "install", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
			}},
			"build": {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
			}},
			"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(web.SPACE, CHROME, CHROME, arg)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
