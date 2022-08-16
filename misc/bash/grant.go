package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const GRANT = "grant"

func init() {
	Index.MergeCommands(ice.Commands{
		GRANT: {Name: "grant hash auto", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			"confirm": {Name: "confirm", Help: "同意", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SESS, mdb.MODIFY, GRANT, m.Option(ice.MSG_USERNAME))
			}},
			"revert": {Name: "revert", Help: "撤销", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SESS, mdb.MODIFY, GRANT, "")
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(SESS, arg); len(arg) > 0 && m.Append(GRANT) == "" {
				m.ProcessConfirm("授权设备")
			}
			m.Tables(func(value ice.Maps) {
				m.PushButton(kit.Select("revert", "confirm", value[GRANT] == ""), mdb.REMOVE)
			})
		}},
	})
}
