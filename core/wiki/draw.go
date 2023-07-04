package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const DRAW = "draw"

func init() {
	Index.MergeCommands(ice.Commands{
		DRAW: {Name: "draw path=src/main.svg@key pid refresh save actions", Help: "思维导图", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, mdb.TYPE, nfs.SVG, mdb.NAME, m.PrefixKey())
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				mdb.IsSearchPreview(m, arg, func() []string { return []string{web.LINK, m.CommandKey(), m.MergePodCmd("", "")} })
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				defer m.Echo("<html><body>").Echo("</body></html>")
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() { m.PushButton(kit.Dict(m.CommandKey(), "导图")) })
			}},
		}, ctx.CmdAction(), WikiAction("", nfs.SVG)), Hand: func(m *ice.Message, arg ...string) {
			kit.If(!_wiki_list(m, arg...), func() {
				_wiki_show(m, arg[0])
				kit.If(m.IsErr(), func() { m.Option(ice.MSG_OUTPUT, "") })
			})
		}},
	})
}
