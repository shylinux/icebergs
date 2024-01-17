package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
)

const SESSION = "session"

func init() {
	Index.MergeCommands(ice.Commands{
		SESSION: {Help: "会话", Role: aaa.VOID, Actions: ice.MergeActions(CmdHashAction(mdb.NAME), mdb.ExportHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.EchoIFrame(m.MergePodCmd("", DESKTOP) + "#" + m.Append(mdb.NAME))
			}
		}},
	})
}
