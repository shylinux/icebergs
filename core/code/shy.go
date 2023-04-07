package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/yac"
	kit "shylinux.com/x/toolkits"
)

const SHY = "shy"

func init() {
	Index.MergeCommands(ice.Commands{
		SHY: {Name: "shy path auto", Help: "笔记", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				// ctx.ProcessCommand(m, web.WIKI_WORD, kit.Simple(path.Join(arg[2], arg[1])))
				// ctx.ProcessCommand(m, yac.STACK, kit.Simple(arg[1]))
				// if ls := kit.Split(arg[1], ice.PS); ls[0] == nfs.SCRIPT {
				// 	m.Search(ls[1], func(key string, cmd *ice.Command) { yac.StackHandler(m) })
				// 	ctx.ProcessCommand(m, ls[1], kit.Simple())
				// } else {
				// 	ctx.ProcessCommand(m, kit.TrimExt(arg[1], SHY), kit.Simple())
				// }
				ctx.ProcessCommand(m, yac.STACK, kit.Simple(arg[1]))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				if msg := m.Cmd(yac.STACK, arg[1]); msg.Option("__index") != "" {
					ctx.ProcessCommand(m, msg.Option("__index"), kit.Simple())
				} else {
					ctx.ProcessCommand(m, yac.STACK, kit.Simple(arg[1]))
				}
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(nfs.Template(m, "demo.shy"), path.Base(path.Dir(path.Join(arg[2], arg[1]))))
			}},
		}, PlugAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == m.CommandKey() {
				m.Cmdy(web.WIKI_WORD, path.Join(ice.SRC, strings.TrimPrefix(arg[0], nfs.SRC)))
			} else {
				m.Cmdy(web.WIKI_WORD, arg)
			}
		}},
	})
}
