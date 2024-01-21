package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const FOOTER = "footer"

func _footer_plugin(m *ice.Message, cmd string, args ice.Any, arg ...string) {
	if p := m.Option(ice.POD); p != "" && arg[0] != cli.RUN {
		defer m.Push(web.SPACE, p)
	}
	if !web.PodCmd(m, ice.POD, kit.Simple(m.ActionKey(), arg)...) {
		ctx.ProcessField(m, cmd, args, arg...)
	}
}
func init() {
	Index.MergeCommands(ice.Commands{
		FOOTER: {Help: "状态栏", Actions: ice.MergeActions(ice.Actions{
			ice.HELP: {Hand: func(m *ice.Message, arg ...string) {
				_footer_plugin(m, web.WIKI_WORD, ctx.GetCmdHelp(m, arg[0]), arg...)
			}},
			nfs.SOURCE: {Hand: func(m *ice.Message, arg ...string) {
				_footer_plugin(m, web.CODE_VIMER, func() []string {
					return nfs.SplitPath(m, ctx.GetCmdFile(m, arg[0]))
				}, arg...)
			}},
			nfs.SCRIPT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SCRIPT, mdb.INSERT, mdb.ZONE, "default", ctx.INDEX, arg[0], ice.AUTO, arg[2])
			}},
			ctx.CONFIG: {Hand: func(m *ice.Message, arg ...string) {
				_footer_plugin(m, ctx.CONFIG, arg, arg...)
			}},
		}, web.ApiWhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Result(kit.Select(ice.Info.Make.Email, mdb.Config(m, TITLE)))
		}},
	})
}
