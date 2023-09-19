package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const ICONS = "icons"

func init() {
	Index.MergeCommands(ice.Commands{
		ICONS: {Actions: ctx.ConfAction(nfs.PATH, "bootstrap-icons/font/bootstrap-icons.css"), Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(lex.SPLIT, ice.USR_MODULES+mdb.Config(m, nfs.PATH), kit.Dict(lex.SPLIT_SPACE, " {:;}"), func(text string, ls []string) {
				if len(ls) > 2 && ls[2] == nfs.CONTENT {
					name := "bi " + strings.TrimPrefix(ls[0], ".")
					m.Push(mdb.NAME, name).Push(mdb.ICON, kit.Format(`<i class="%s"></i>`, name))
				}
			})
			m.StatusTimeCount()
		}},
	})
}
