package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const SESSION = "session"

func init() {
	Index.MergeCommands(ice.Commands{
		SESSION: {Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { mdb.HashImport(m) }},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { mdb.HashExport(m) }},
		}, CmdHashAction(mdb.NAME), mdb.ImportantHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.EchoIFrame(m.MergePodCmd("", DESKTOP, SESSION, arg[0]))
			}
		}},
	})
}
