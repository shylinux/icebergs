package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"
)

const SCAN = "scan"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SCAN: {Name: SCAN, Help: "扫码", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			SCAN: {Name: "scan hash auto create@scanQRCode scanQRCode0=应用扫码", Help: "扫码", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=text name=hi text:textarea=hi", Help: "扫码", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, SCAN, "", mdb.HASH, arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SCAN, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, SCAN, "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, SCAN, "", mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, SCAN, "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					text := m.Cmd(mdb.SELECT, SCAN, "", mdb.HASH, kit.MDB_HASH, arg[0]).Append(kit.MDB_TEXT)
					m.Cmdy(wiki.IMAGE, "qrcode", text)
					m.Cmdy(wiki.SPARK, "inner", text)
					m.Render("")
					return
				}

				m.Cmdy(mdb.SELECT, SCAN, "", mdb.HASH)
				m.SortTimeR(kit.MDB_TIME)
				m.PushAction(mdb.REMOVE)
			}},
		},
	}, nil)
}
