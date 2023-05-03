package macos

import (
	ice "shylinux.com/x/icebergs"
)

const DOCK = "dock"

func init() { Index.MergeCommands(ice.Commands{DOCK: {Actions: CmdHashAction()}}) }

func DockAppend(m *ice.Message, name, index string, arg ...string) {
	install(m, DOCK, name, index, arg...)
}
