package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
)

const DRAW = "draw"

func init() {
	Index.MergeCommands(ice.Commands{
		DRAW: {Name: "draw path=src/main.svg pid list save actions", Help: "思维导图", Icon: "Grapher.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, mdb.Config(m, lex.REGEXP), m.ShortKey())
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				defer m.Echo("<!DOCTYPE html><html><body>").Echo("</body></html>")
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
		}, WikiAction("", nfs.SVG))},
	})
}
