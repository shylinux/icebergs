package macos

import ice "shylinux.com/x/icebergs"

const MENU = "menu"

func init() { Index.MergeCommands(ice.Commands{MENU: {Actions: CmdHashAction()}}) }
