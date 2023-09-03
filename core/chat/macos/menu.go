package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
)

const MENU = "menu"

func init() {
	Index.MergeCommands(ice.Commands{
		MENU: {Actions: ice.MergeActions(aaa.RoleAction(), CmdHashAction(), mdb.ImportantHashAction())},
	})
}
func MenuAppend(m *ice.Message, icon, index string) {
	install(m, MENU, icon, index)
}
