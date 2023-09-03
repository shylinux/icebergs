package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _makefile_xterm(m *ice.Message, arg ...string) {
	ctx.Process(m, XTERM, func() []string {
		m.Push(ctx.STYLE, html.OUTPUT)
		if ls := kit.Simple(kit.UnMarshal(m.Option(ctx.ARGS))); len(ls) > 0 {
			return ls
		}
		return []string{mdb.TYPE, "sh", nfs.PATH, kit.Select("", kit.Dir(arg[2], arg[1]), arg[2] != ice.SRC)}
	}, arg...)
}

const MAKEFILE = "makefile"

func init() {
	Index.MergeCommands(ice.Commands{
		MAKEFILE: {Name: "makefile path auto", Help: "构建", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Options(lex.SPLIT_BLOCK, ":")
				m.Cmd(lex.SPLIT, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), func(indent int, text string, ls []string) {
					if indent == 1 && ls[1] == ":" {
						m.Push("target", ls[0]).Push("source", kit.Join(ls[2:])).PushButton("make")
					}
				})
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _makefile_xterm(m, arg...) }},
		}, PlugAction())},
	})
}
