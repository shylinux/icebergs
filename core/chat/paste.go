package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
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
			PASTE: {Name: "paste hash auto getClipboardData", Help: "粘贴板", Action: map[string]*ice.Action{
				"getClipboardData": {Name: "getClipboardData", Help: "粘贴", Hand: func(m *ice.Message, arg ...string) {
					_trans(arg, map[string]string{"data": "text"})
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), "", mdb.HASH, arg)
				}},
				mdb.CREATE: {Name: "create type=text name=hi data:textarea=hi", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_trans(arg, map[string]string{"data": "text"})
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(PASTE), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(PASTE), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(PASTE), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(PASTE), "", mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, m.Prefix(PASTE), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(PASTE, kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, cmd, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) > 0 {
					m.PushScript("script", m.Append(kit.MDB_TEXT))
					m.PushQRCode("qrcode", m.Append(kit.MDB_TEXT))
				}
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
