package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const REWRITE = "rewrite"

func init() {
	Index.MergeCommands(ice.Commands{
		REWRITE: {Actions: ice.MergeActions(ice.Actions{}, mdb.HashAction(mdb.SHORT, "space,index", mdb.FIELD, "time,hash,space,index,to_space,to_index"))},
	})
}
