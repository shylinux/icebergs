package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
)

const IFRAME = "iframe"

func init() {
	Index.MergeCommands(ice.Commands{
		IFRAME: {Name: "iframe hash auto", Help: "网页", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create link name type", Help: "创建"},
		}, mdb.HashAction(mdb.SHORT, mdb.LINK, mdb.FIELD, "time,hash,type,name,link")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || arg[0] == "" {
				m.Action(mdb.CREATE, mdb.PRUNES)
			} else {
				m.StatusTime(mdb.LINK, m.Append(mdb.LINK))
				m.Action(cli.OPEN)
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}
