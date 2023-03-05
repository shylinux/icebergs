package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const SHY = "shy"

func init() {
	Index.MergeCommands(ice.Commands{
		SHY: {Name: "shy path auto", Help: "脚本", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessCommand(m, web.WIKI_WORD, kit.Simple(path.Join(arg[2], arg[1])))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessCommand(m, web.WIKI_WORD, kit.Simple(path.Join(arg[2], arg[1])))
				// m.Cmdy(ssh.SOURCE, path.Join(arg[2], arg[1]))
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) { m.Echo(`chapter "hi"`) }},
		}, PlugAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == m.CommandKey() {
				m.Cmdy(web.WIKI_WORD, path.Join(ice.SRC, strings.TrimPrefix(arg[0], "src/")))
			} else {
				m.Cmdy(web.WIKI_WORD, arg)
			}
		}},
	})
}
