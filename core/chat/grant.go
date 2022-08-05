package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const GRANT = "grant"

func init() {
	Index.MergeCommands(ice.Commands{
		GRANT: {Name: "grant space id auto insert", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case web.SPACE:
					m.Cmdy(web.SPACE).RenameAppend(mdb.NAME, web.SPACE).Cut("space,type")
				case GRANT:
					m.Cmdy(web.SPACE, m.Option(web.SPACE), web.SPACE).RenameAppend(mdb.NAME, GRANT).Cut("grant,type")
				case aaa.USERROLE:
					m.Push(arg[0], m.Option(ice.MSG_USERROLE))
				case aaa.USERNAME:
					m.Push(arg[0], m.Option(ice.MSG_USERNAME))
				}
			}},
			mdb.INSERT: {Name: "insert space grant userrole username", Help: "添加"},
		}, mdb.ZoneAction(mdb.SHORT, web.SPACE, mdb.FIELD, "time,grant,userrole,username"))},
	})
}
