package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
)

const SUPPLY = "supply"

func init() {
	Index.MergeCommands(ice.Commands{
		SUPPLY: {Help: "供应商", Actions: ice.MergeActions(mdb.HashAction(
			mdb.SHORT, aaa.USERNAME,
		)), Hand: func(m *ice.Message, arg ...string) {

		}},
	})
}
