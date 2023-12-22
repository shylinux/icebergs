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
		FOOTER: {Help: "状态栏", Actions: ice.MergeActions(ice.Actions{
			ice.HELP: {Hand: func(m *ice.Message, arg ...string) {
				if !web.PodCmd(m, ice.POD, kit.Simple(m.ActionKey(), arg)...) {
					ctx.ProcessField(m, web.WIKI_WORD, ctx.GetCmdHelp(m, arg[0]), arg...)
				}
			}},
			nfs.SOURCE: {Hand: func(m *ice.Message, arg ...string) {
				if !web.PodCmd(m, ice.POD, kit.Simple(m.ActionKey(), arg)...) {
					ctx.ProcessField(m, web.CODE_INNER, func() []string { return nfs.SplitPath(m, ctx.GetCmdFile(m, arg[0])) }, arg...)
				}
			}},
			ctx.CONFIG: {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessField(m, ctx.CONFIG, arg, arg...) }},
		}, web.ApiWhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Result(kit.Select(ice.Info.Make.Email, mdb.Config(m, TITLE)))
		}},
	})
}
