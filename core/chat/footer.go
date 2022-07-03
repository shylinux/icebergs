package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	LEGAL = "legal"
)
const FOOTER = "footer"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FOOTER: {Name: FOOTER, Help: "状态栏", Value: kit.Dict(LEGAL, kit.List(`<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`))},
	}, Commands: map[string]*ice.Command{
		web.P(FOOTER): {Name: "/footer", Help: "状态栏", Action: ice.MergeAction(map[string]*ice.Action{
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(arg); m.Result(1) == ice.ErrNotFound {
					m.Set(ice.MSG_RESULT).Cmdy(cli.SYSTEM, arg)
				}
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Confm(FOOTER, LEGAL, func(index int, value string) { m.Echo(value) })
		}},
	}})
}
