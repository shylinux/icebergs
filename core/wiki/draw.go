package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DRAW = "draw"

func init() {
	Index.MergeCommands(ice.Commands{
		DRAW: {Name: "draw path=src/main.svg pid refresh:button=auto save edit actions", Help: "思维导图", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, mdb.TYPE, nfs.SVG, mdb.NAME, m.PrefixKey())
			}},
			mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				m.Echo("<html><body>")
				defer m.Echo("</body></html>")
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
			nfs.SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				_wiki_save(m, arg[0], m.Option(nfs.CONTENT))
			}},
		}, WikiAction("", nfs.SVG), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if !_wiki_list(m, kit.Select(nfs.PWD, arg, 0)) {
				_wiki_show(m, arg[0])
			}
		}},
	})
}
