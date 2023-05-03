package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const SESSION = "session"

func init() { Index.MergeCommands(ice.Commands{SESSION: {Actions: CmdHashAction(mdb.NAME)}}) }
