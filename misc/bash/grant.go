package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const GRANT = "grant"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		"grant": {Name: "grant hash auto", Help: "授权", Actions: ice.MergeAction(ice.Actions{
			"confirm": {Name: "confirm", Help: "同意", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SESS, mdb.MODIFY, GRANT, m.Option(ice.MSG_USERNAME), ice.Option{mdb.HASH, m.Option("hash")})
			}},
			"revert": {Name: "confirm", Help: "撤销", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SESS, mdb.MODIFY, GRANT, "", ice.Option{mdb.HASH, m.Option("hash")})
			}},
			"remove": {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SESS, mdb.REMOVE, mdb.HASH, m.Option(mdb.HASH))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(SESS, arg); len(arg) > 0 && m.Append("grant") == "" {
				m.Process("_confirm", "授权设备")
			}
			m.Tables(func(value ice.Maps) {
				if value["grant"] == "" {
					m.PushButton("confirm", mdb.REMOVE)
				} else {
					m.PushButton("revert", mdb.REMOVE)
				}
			})
		}},
	}})
}
