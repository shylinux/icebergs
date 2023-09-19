package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const FOOTER = "footer"

func init() {
	Index.MergeCommands(ice.Commands{
		FOOTER: {Actions: ice.MergeActions(ice.Actions{
			ice.HELP: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, web.WIKI_WORD, []string{ice.SRC_DOCUMENT + arg[0] + "/list.shy"}, arg...)
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
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if aaa.Right(m, arg) {
					if m.Cmdy(arg); m.IsErrNotFound() {
						m.RenderResult(m.Cmdx(cli.SYSTEM, arg))
					}
				}
			}},
		}, web.ApiAction(), aaa.WhiteAction(ctx.RUN, ctx.COMMAND), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Result(kit.Select(ice.Info.Make.Email, mdb.Config(m, TITLE)))
		}},
	})
}
