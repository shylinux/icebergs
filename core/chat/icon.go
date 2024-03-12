package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const ICON = "icon"

func init() {
	Index.MergeCommands(ice.Commands{
		ICON: {Help: "图标", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				kit.For([]string{
					"bootstrap-icons/font/fonts/bootstrap-icons.woff2",
					"bootstrap-icons/font/bootstrap-icons.css",
				}, func(p string) {
					// m.Cmd(WEBPACK, mdb.INSERT, p)
					m.Cmd(web.BINPACK, mdb.INSERT, nfs.USR_MODULES+p)
				})
			}},
		}, ctx.ConfAction(nfs.PATH, "bootstrap-icons/font/bootstrap-icons.css")), Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(lex.SPLIT, nfs.USR_MODULES+mdb.Config(m, nfs.PATH), kit.Dict(lex.SPLIT_SPACE, " {:;}"), func(text string, ls []string) {
				if len(ls) > 2 && ls[2] == nfs.CONTENT {
					name := "bi " + strings.TrimPrefix(ls[0], nfs.PT)
					m.Push(mdb.NAME, name).Push(mdb.ICON, kit.Format(`<i class="%s"></i>`, name))
				}
			})
		}},
	})
}
