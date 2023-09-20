package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
)

const SEARCH = "search"

func init() {
	Index.MergeCommands(ice.Commands{
		SEARCH: {Actions: ice.MergeActions(ice.Actions{
			cli.OPENS: {Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, arg...) }},
		}, web.ApiAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(web.Space(m, m.Option(ice.POD)), "mdb.search", arg).StatusTimeCount()
		}},
	})
}
