package macos

import ice "shylinux.com/x/icebergs"

const DESKTOP = "desktop"

func init() { Index.MergeCommands(ice.Commands{DESKTOP: {Actions: CmdHashAction()}}) }

func DeskAppend(m *ice.Message, name, index string, arg ...string) {
	install(m, DESKTOP, name, index, arg...)
}
