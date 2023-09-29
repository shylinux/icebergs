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
				ctx.ProcessField(m, web.WIKI_WORD, kit.Simple(path.Join(arg[2], arg[1])))
				return
				ctx.ProcessField(m, yac.STACK, kit.Simple(path.Join(arg[2], arg[1])))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, web.WIKI_WORD, kit.Simple(path.Join(arg[2], arg[1])))
				return
				if msg := m.Cmd(yac.STACK, path.Join(arg[2], arg[1])); msg.Option("__index") != "" {
					ctx.ProcessField(m, msg.Option("__index"), kit.Simple())
				} else {
					ctx.ProcessField(m, yac.STACK, kit.Simple(path.Join(arg[2], arg[1])))
				}
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.NAME, path.Base(path.Dir(path.Join(arg[2], arg[1]))))
				m.Echo(nfs.Template(m, "demo.shy"))
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.wiki.word", COMPLETE, arg) }},
		}, PlugAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == m.CommandKey() {
				m.Cmdy(web.WIKI_WORD, path.Join(ice.SRC, strings.TrimPrefix(arg[0], nfs.SRC)))
			} else {
				m.Cmdy(web.WIKI_WORD, arg)
			}
		}},
	})
}
