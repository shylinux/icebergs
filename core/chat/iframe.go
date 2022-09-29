package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const IFRAME = "iframe"

func init() {
	Index.MergeCommands(ice.Commands{
		IFRAME: {Name: "iframe hash auto", Help: "浏览器", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create link name type", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.NAME, kit.ParseURL(m.Option(mdb.LINK)).Host, mdb.TYPE, mdb.LINK)
				mdb.HashCreate(m, m.OptionSimple("link,name,type"))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.LINK, mdb.FIELD, "time,hash,type,name,link")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || arg[0] == "" {
				m.Action(mdb.CREATE, mdb.PRUNES)
			} else {
				if m.Length() == 0 {
					m.Append(mdb.LINK, arg[0])
				}
				m.Action(cli.OPEN).StatusTime(mdb.LINK, m.Append(mdb.LINK))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}
