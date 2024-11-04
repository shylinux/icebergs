package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const corporation = "corporation"
	Index.MergeCommands(ice.Commands{
		corporation: {Name: "corporation username auto", Help: "法人", Meta: kit.Dict(
			"_trans.input", kit.Dict(
				"idnumber", "身份证号",
				"usci", "统一社会信用代码",
				"account", "对公账户",
				"bank", "开户银行",
			),
		), Actions: ice.MergeActions(ice.Actions{}, mdb.ExportHashAction(
			mdb.SHORT, "username", mdb.FIELD, "time,username,mobile,idnumber,usci,account,bank,email,portal",
		)), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Action(mdb.CREATE)
			web.PushPodCmd(m, "", arg...)
		}},
	})
}
