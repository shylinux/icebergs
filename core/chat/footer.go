package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const FOOTER = "footer"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(FOOTER): {Name: "/footer", Help: "状态栏", Actions: ice.MergeActions(ice.Actions{
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if aaa.Right(m, arg) {
					if m.Cmdy(arg); m.IsErrNotFound() {
						m.RenderResult(m.Cmdx(cli.SYSTEM, arg))
					}
				}
			}},
			ice.HELP: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, web.WIKI_WORD, func() string { return kit.ExtChange(ctx.GetCmdFile(m, arg[0]), nfs.SHY) }, arg...)
			}},
			nfs.SCRIPT: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, web.CODE_VIMER, func() []string {
					return nfs.SplitPath(m, kit.ExtChange(nfs.Relative(m, ctx.GetCmdFile(m, arg[0])), nfs.JS))
				}, arg...)
			}},
			nfs.SOURCE: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, web.CODE_VIMER, func() []string { return nfs.SplitPath(m, ctx.GetCmdFile(m, arg[0])) }, arg...)
			}},
			ctx.CONFIG: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, ctx.CONFIG, arg, arg...)
			}},
		}, ctx.CmdAction(), aaa.WhiteAction(ctx.COMMAND, ice.RUN)), Hand: func(m *ice.Message, arg ...string) {
			m.Result(kit.Select(m.Config(TITLE), ice.Info.Make.Email))
		}},
	})
}
