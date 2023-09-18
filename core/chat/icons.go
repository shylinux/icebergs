package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const ICONS = "icons"
	Index.MergeCommands(ice.Commands{
		ICONS: {Hand: func(m *ice.Message, arg ...string) {
			m.Option(lex.SPLIT_SPACE, " {:;}")
			m.Cmd(lex.SPLIT, "usr/node_modules/bootstrap-icons/font/bootstrap-icons.css", func(text string, ls []string) {
				if len(ls) > 2 && ls[2] == "content" {
					name := "bi " + strings.TrimPrefix(ls[0], ".")
					m.Push("name", name)
					m.Push("icon", kit.Format(`<i class="%s"></i>`, name))
				}
			})
			m.StatusTimeCount()
		}},
	})
}
