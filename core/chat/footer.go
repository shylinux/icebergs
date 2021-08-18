package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	kit "shylinux.com/x/toolkits"
)

const (
	LEGAL = "legal"
)
const FOOTER = "footer"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FOOTER: {Name: FOOTER, Help: "状态栏", Value: kit.Dict(
				LEGAL, []interface{}{`<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`},
			)},
		},
		Commands: map[string]*ice.Command{
			"/footer": {Name: "/footer", Help: "状态栏", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(ctx.COMMAND, arg)
				}},
				cli.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				kit.Fetch(m.Confv(FOOTER, LEGAL), func(index int, value string) { m.Echo(value) })
			}},
		},
	})
}
