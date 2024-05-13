package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
)

const TUTOR = "tutor"

func init() {
	Index.MergeCommands(ice.Commands{
		TUTOR: {Name: "tutor zone id auto", Help: "向导", Actions: mdb.ZoneAction(
			mdb.SHORT, "zone", mdb.FIELD, "time,zone", mdb.FIELDS, "time,id,type,name,text",
		), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Push(mdb.TIME, m.Time()).Push(mdb.ZONE, "_current")
			} else {
				m.Action(cli.PLAY)
			}
			m.Display("")
		}},
	})
}
