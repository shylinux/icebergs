package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/ssh"
)

const PASTE = "paste"

func init() {
	Index.MergeCommands(ice.Commands{
		PASTE: {Name: "paste hash auto getClipboardData", Help: "粘贴", Actions: ice.MergeAction(ice.Actions{
			"getClipboardData": {Name: "getClipboardData", Help: "粘贴", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(PASTE, mdb.CREATE, arg)
			}},
			mdb.CREATE: {Name: "create type=text name=hi text:textarea=hi", Help: "添加"},
		}, mdb.HashAction(mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.PushScript(ssh.SCRIPT, m.Append(mdb.TEXT))
				m.PushQRCode(cli.QRCODE, m.Append(mdb.TEXT))
			}
		}},
	})
}
