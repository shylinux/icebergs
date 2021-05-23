package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	kit "github.com/shylinux/toolkits"
)

const (
	LEGAL = "legal"
)
const P_FOOTER = "/footer"
const FOOTER = "footer"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FOOTER: {Name: FOOTER, Help: "状态栏", Value: kit.Dict(
				LEGAL, []interface{}{`<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`},
			)},
		},
		Commands: map[string]*ice.Command{
			P_FOOTER: {Name: "/footer", Help: "状态栏", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) > 0 && arg[0] == "run" {
						if m.Right(arg[1:]) {
							m.Cmdy(arg[1:])
						}
						return
					}
					m.Cmdy(ctx.COMMAND, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				kit.Fetch(m.Confv(FOOTER, LEGAL), func(index int, value string) { m.Echo(value) })
			}},
		},
	})
}
