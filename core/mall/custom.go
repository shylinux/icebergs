package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
)

const CUSTOM = "custom"

func init() {
	Index.MergeCommands(ice.Commands{
		CUSTOM: {Help: "顾客", Actions: ice.MergeActions(mdb.HashAction(
			mdb.SHORT, aaa.USERNAME,
		)), Hand: func(m *ice.Message, arg ...string) {

		}},
	})
}
