package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const SHY = "shy"

func init() {
	Index.MergeCommands(ice.Commands{
		SHY: {Name: "shy path auto", Help: "笔记", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessCommand(m, web.WIKI_WORD, kit.Simple(path.Join(arg[2], arg[1])))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessCommand(m, web.WIKI_WORD, kit.Simple(path.Join(arg[2], arg[1])))
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_shy_template, path.Base(path.Dir(path.Join(arg[2], arg[1]))))
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

var _shy_template = `chapter "%s"`
