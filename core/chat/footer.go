package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
)

const (
	EMAIL = "email"
)
const FOOTER = "footer"

func init() {
	Index.MergeCommands(ice.Commands{
		FOOTER: {Name: "footer", Help: "状态栏", Actions: ice.MergeActions(ice.Actions{
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				if aaa.Right(m, arg) {
					if m.Cmdy(arg); m.IsErrNotFound() {
						m.SetResult().Cmdy(cli.SYSTEM, arg)
					}
				}
			}},
		}, ctx.CmdAction(EMAIL, `<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`), web.ApiAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Echo(m.Config(EMAIL))
		}},
	})
}
