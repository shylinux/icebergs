package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const SCRIPT = "script"

func init() {
	Index.MergeCommands(ice.Commands{
		SCRIPT: {Name: "script zone id auto", Help: "脚本", Actions: ice.MergeActions(mdb.ZoneAction(mdb.FIELDS, "time,index,auto")), Hand: func(m *ice.Message, arg ...string) {
			mdb.ZoneSelect(m, arg...)
			m.Display("")
			if len(arg) == 0 {
				m.PushAction("play", mdb.REMOVE)
			} else {
				m.Action("play")
			}
		}},
	})
}
