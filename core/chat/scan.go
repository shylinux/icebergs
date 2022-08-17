package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/ssh"
)

const SCAN = "scan"

func init() {
	Index.MergeCommands(ice.Commands{
		SCAN: {Name: "scan hash auto scanQRCode scanQRCode0", Help: "扫码", Actions: ice.MergeActions(ice.Actions{
			"scanQRCode0": {Name: "scan create", Help: "本机扫码"},
			"scanQRCode":  {Name: "scan create", Help: "扫码"},
			mdb.CREATE:    {Name: "create type=text name=hi text:textarea=hi", Help: "添加"},
		}, mdb.HashAction(mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				if m.Append(mdb.TYPE) == "image" {
					m.PushImages("image", m.Append(mdb.TEXT))
				}
				m.PushScript(ssh.SCRIPT, m.Append(mdb.TEXT))
				m.PushQRCode(cli.QRCODE, m.Append(mdb.TEXT))
			}
		}},
	})
}
