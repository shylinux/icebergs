package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"
)

const PASTE = "paste"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PASTE: {Name: PASTE, Help: "粘贴板", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			PASTE: {Name: "paste hash auto 添加 导出 导入", Help: "粘贴板", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert text:textarea=hi", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, arg)
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					text := m.Cmd(mdb.SELECT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_HASH, arg[0]).Append(kit.MDB_TEXT)
					m.Cmdy(wiki.SPARK, "inner", text)
					m.Cmdy(wiki.IMAGE, "qrcode", text)
					m.Render("")
					return
				}

				m.Cmdy(mdb.SELECT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				m.Sort(kit.MDB_TIME, "time_r")
				m.PushAction("删除")
			}},
		},
	}, nil)
}
