package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
)

const FOOTER = "footer"

func init() {
	const EMAIL = "email"
	Index.MergeCommands(ice.Commands{
		web.P(FOOTER): {Name: "/footer", Help: "状态栏", Actions: ice.MergeActions(ice.Actions{
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if aaa.Right(m, arg) {
					if m.Cmdy(arg); m.IsErrNotFound() {
						m.SetResult().Cmdy(cli.SYSTEM, arg)
					}
				}
			}},
		}, ctx.CmdAction(EMAIL, `shylinuxc@gmail.com`), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Result(m.Configv(EMAIL))
		}},
	})
}
