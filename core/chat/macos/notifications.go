package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const NOTIFICATIONS = "notifications"

func init() {
	Index.MergeCommands(ice.Commands{NOTIFICATIONS: {Name: "notifications list", Actions: ice.MergeActions(ice.Actions{
		mdb.PRUNES: {Name: "prunes", Hand: func(m *ice.Message, arg ...string) {
			m.Conf("", kit.Keys(mdb.HASH), "")
		}},
	}, CmdHashAction()), Hand: func(m *ice.Message, arg ...string) {
		mdb.HashSelect(m, arg...).SortStrR(mdb.TIME).Display("")
	}}})
}
func Notify(m *ice.Message, name, text string, arg ...string) {
	m.Cmd(NOTIFICATIONS, mdb.CREATE, mdb.NAME, name, mdb.TEXT, text, arg)
}
