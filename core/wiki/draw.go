package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DRAW = "draw"

func init() {
	Index.MergeCommands(ice.Commands{
		DRAW: {Name: "draw path=src/main.svg pid list save actions", Icon: "Grapher.png", Help: "思维导图", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, mdb.TYPE, nfs.SVG, mdb.NAME, m.PrefixKey())
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				defer m.Echo("<!DOCTYPE html><html><body>").Echo("</body></html>")
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
		}, aaa.RoleAction(), WikiAction("", nfs.SVG)), Hand: func(m *ice.Message, arg ...string) {
			kit.If(!_wiki_list(m, arg...), func() {
				_wiki_show(m, arg[0])
				kit.If(m.IsErr(), func() { m.Option(ice.MSG_OUTPUT, "") })
			})
		}},
	})
}
