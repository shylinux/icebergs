package chat

import (
	ice "shylinux.com/x/icebergs"
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
				ctx.ProcessField(m, web.WIKI_WORD, []string{ctx.FileURI(kit.ExtChange(ctx.GetCmdFile(m, arg[0]), nfs.SHY))}, arg...)
			}},
			nfs.SCRIPT: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, web.CODE_VIMER, func() []string {
					return nfs.SplitPath(m, kit.ExtChange(nfs.Relative(m, ctx.GetCmdFile(m, arg[0])), nfs.JS))
				}, arg...)
			}},
			nfs.SOURCE: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, web.CODE_VIMER, func() []string { return nfs.SplitPath(m, ctx.GetCmdFile(m, arg[0])) }, arg...)
			}},
			ctx.CONFIG: {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessField(m, ctx.CONFIG, arg, arg...) }},
		}, web.ApiWhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Result(kit.Select(ice.Info.Make.Email, mdb.Config(m, TITLE)))
		}},
	})
}
