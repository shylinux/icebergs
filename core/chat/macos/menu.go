package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const MENU = "menu"

func init() {
	Index.MergeCommands(ice.Commands{MENU: {Actions: ice.MergeActions(ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			if m.Cmd(MENU).Length() == 0 {
				MenuAppend(m, "", Prefix(NOTIFICATIONS))
				MenuAppend(m, "", Prefix(SEARCHS))
			}
		}},
	}, mdb.ImportantHashAction(), CmdHashAction())}})
}
func MenuAppend(m *ice.Message, name, index string, arg ...string) {
	install(m, MENU, name, index, arg...)
}
