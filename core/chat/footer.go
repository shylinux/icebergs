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
		web.P(FOOTER): {Name: "footer", Help: "状态栏", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.CommandKey())
			}},
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if aaa.Right(m, arg) {
					if m.Cmdy(arg); m.IsErrNotFound() {
						m.SetResult().Cmdy(cli.SYSTEM, arg)
					}
				}
			}},
		}, ctx.CmdAction(EMAIL, `<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`)), Hand: func(m *ice.Message, arg ...string) {
			m.Result(m.Configv(EMAIL))
		}},
	})
}
