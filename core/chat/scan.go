package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const SCAN = "scan"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SCAN: {Name: SCAN, Help: "二维码", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_TEXT, kit.MDB_FIELD, "time,hash,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			SCAN: {Name: "scan hash auto scanQRCode scanQRCode0", Help: "二维码", Action: map[string]*ice.Action{
				"scanQRCode0": {Name: "create", Help: "本机扫码", Hand: func(m *ice.Message, arg ...string) {}},
				"scanQRCode": {Name: "create", Help: "扫码", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(SCAN), "", mdb.HASH, arg)
				}},
				mdb.CREATE: {Name: "create type=text name=hi text:textarea=hi", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(SCAN), "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(SCAN), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(SCAN), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(SCAN), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(SCAN), "", mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, m.Prefix(SCAN), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(SCAN, kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, m.Prefix(SCAN), "", mdb.HASH, kit.MDB_HASH, arg); len(arg) > 0 {
					m.PushScript("script", m.Append(kit.MDB_TEXT))
					m.PushQRCode("qrcode", m.Append(kit.MDB_TEXT))
				}
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
