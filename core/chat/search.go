package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const SEARCH = "search"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(SEARCH): {Name: "/search", Help: "搜索框", Actions: ice.MergeActions(ice.Actions{
			cli.OPENS: {Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, arg...) }},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(web.Space(m, m.Option(ice.POD)), mdb.SEARCH, arg).StatusTimeCount()
		}},
	})
}
