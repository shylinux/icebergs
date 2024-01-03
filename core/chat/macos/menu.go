package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
)

const MENU = "menu"

func init() {
	Index.MergeCommands(ice.Commands{
		MENU: {Help: "菜单栏", Role: aaa.VOID, Actions: ice.MergeActions(CmdHashAction(), mdb.ClearOnExitHashAction())},
	})
}

func MenuAppend(m *ice.Message, icon, index string) { install(m, MENU, icon, index) }
