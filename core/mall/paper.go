package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const PAPER = "paper"

func init() {
	Index.MergeCommands(ice.Commands{
		PAPER: {Name: "paper", Help: "问卷", Actions: ice.MergeActions(ice.Actions{}, mdb.ZoneAction(mdb.FIELD, "time,id,type,name,text"))},
	})
}
