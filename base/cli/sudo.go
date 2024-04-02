package cli

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

func init() {
	Index.MergeCommands(ice.Commands{
		SUDO: {Actions: mdb.HashAction(mdb.SHORT, "cmd", mdb.FIELD, "time,cmd")},
	})
}
