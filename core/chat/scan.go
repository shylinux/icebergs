package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/ssh"
	kit "shylinux.com/x/toolkits"
)

const SCAN = "scan"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SCAN: {Name: SCAN, Help: "扫码", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_TEXT, kit.MDB_FIELD, "time,hash,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		SCAN: {Name: "scan hash auto scanQRCode scanQRCode0", Help: "扫码", Meta: kit.Dict(
			ice.Display("scan.js"),
		), Action: ice.MergeAction(map[string]*ice.Action{
			"scanQRCode0": {Name: "scan create", Help: "本机扫码"},
			"scanQRCode":  {Name: "scan create", Help: "扫码"},
			mdb.CREATE:    {Name: "create type=text name=hi text:textarea=hi", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.PushScript(ssh.SCRIPT, m.Append(kit.MDB_TEXT))
				m.PushQRCode(cli.QRCODE, m.Append(kit.MDB_TEXT))
			}
		}},
	}})
}
