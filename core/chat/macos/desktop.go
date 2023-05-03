package macos

import ice "shylinux.com/x/icebergs"

const DESKTOP = "desktop"

func init() { Index.MergeCommands(ice.Commands{DESKTOP: {Actions: CmdHashAction()}}) }
