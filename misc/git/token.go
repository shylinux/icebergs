package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
)

const TOKEN = "token"

func init() {
	Index.MergeCommands(ice.Commands{
		TOKEN: {Name: "token username auto", Actions: ice.MergeActions(mdb.HashAction(
			mdb.SHORT, aaa.USERNAME, mdb.EXPIRE, "720h", mdb.FIELD, "time,username,token",
		))},
	})
}
