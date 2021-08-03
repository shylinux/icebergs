package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/ssh"
	kit "github.com/shylinux/toolkits"
)

const PASTE = "paste"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PASTE: {Name: PASTE, Help: "粘贴板", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_TEXT, kit.MDB_FIELD, "time,hash,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			PASTE: {Name: "paste hash auto getClipboardData", Help: "粘贴板", Action: ice.MergeAction(map[string]*ice.Action{
				"getClipboardData": {Name: "getClipboardData", Help: "粘贴", Hand: func(m *ice.Message, arg ...string) {
					_trans(arg, map[string]string{"data": "text"})
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), "", mdb.HASH, arg)
				}},
				mdb.CREATE: {Name: "create type=text name=hi data:textarea=hi", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_trans(arg, map[string]string{"data": "text"})
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), "", mdb.HASH, arg)
				}},
			}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(PASTE, kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, cmd, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) > 0 {
					m.PushScript(ssh.SCRIPT, m.Append(kit.MDB_TEXT))
					m.PushQRCode(cli.QRCODE, m.Append(kit.MDB_TEXT))
				}
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
