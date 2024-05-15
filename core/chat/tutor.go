package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

const TUTOR = "tutor"

func init() {
	Index.MergeCommands(ice.Commands{
		TUTOR: {Name: "tutor zone id auto", Help: "向导", Actions: ice.MergeActions(ice.Actions{
			nfs.SAVE: {Name: "save zone*", Hand: func(m *ice.Message, arg ...string) {}},
		}, mdb.ZoneAction(
			mdb.SHORT, "zone", mdb.FIELD, "time,zone,count", mdb.FIELDS, "time,id,type,text",
		)), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelectAll(m, arg...); len(arg) == 0 {
				m.Push(mdb.TIME, m.Time()).Push(mdb.ZONE, "_current")
			} else if m.SortInt(mdb.ID); arg[0] == "_current" {
				m.Action(web.PLAY, nfs.SAVE)
			} else {
				m.PushAction(web.SHOW, mdb.VIEW, mdb.DATA).Action(web.PLAY)
			}
			m.Display("").DisplayCSS("")
		}},
	})
}
