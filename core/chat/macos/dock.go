package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
)

const DOCK = "dock"

func init() {
	Index.MergeCommands(ice.Commands{
		DOCK: {Help: "工具栏", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{}, CmdHashAction(), mdb.ExportHashAction())},
	})
}

func DockAppend(m *ice.Message, icon, index string) { install(m, DOCK, icon, index) }
