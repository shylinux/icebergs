package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _makefile_xterm(m *ice.Message, arg ...string) {
	ctx.Process(m, XTERM, func() []string {
		m.Push(ctx.STYLE, html.OUTPUT)
		return []string{mdb.TYPE, "sh", nfs.PATH, kit.Select("", kit.Dir(arg[2], arg[1]), arg[2] != ice.SRC)}
	}, arg...)
}

const MAKEFILE = "makefile"

func init() {
	Index.MergeCommands(ice.Commands{
		MAKEFILE: {Name: "makefile path auto", Help: "构建", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _makefile_xterm(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _makefile_xterm(m, arg...) }},
		}, PlugAction())},
	})
}
