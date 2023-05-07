package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const SESSION = "session"

func init() {
	Index.MergeCommands(ice.Commands{SESSION: {Actions: ice.MergeActions(mdb.ImportantHashAction(), CmdHashAction(mdb.NAME)), Hand: func(m *ice.Message, arg ...string) {
		if mdb.HashSelect(m, arg...); len(arg) > 0 {
			m.EchoIFrame(m.MergePodCmd("", DESKTOP, SESSION, arg[0]))
		}
	}}})
}
