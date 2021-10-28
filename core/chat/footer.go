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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FOOTER: {Name: FOOTER, Help: "状态栏", Value: kit.Dict(
			LEGAL, kit.List(`<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`),
		)},
	}, Commands: map[string]*ice.Command{
		"/footer": {Name: "/footer", Help: "状态栏", Action: map[string]*ice.Action{
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(ctx.COMMAND, arg)
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(arg); m.Result(1) == ice.ErrNotFound {
					m.Set(ice.MSG_RESULT).Cmdy(cli.SYSTEM, arg)
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Confm(FOOTER, LEGAL, func(index int, value string) { m.Echo(value) })
		}},
	}})
}
