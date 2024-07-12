package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

func init() {
	const corporation = "corporation"
	Index.MergeCommands(ice.Commands{
		corporation: {Name: "corporation username auto", Help: "法人", Actions: ice.MergeActions(ice.Actions{}, mdb.ExportHashAction(
			mdb.SHORT, "username", mdb.FIELD, "time,username,mobile,idnumber,usci,account,bank,email,portal",
		)), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
			web.PushPodCmd(m, "", arg...)
			m.Action(mdb.CREATE)
		}},
	})
}
