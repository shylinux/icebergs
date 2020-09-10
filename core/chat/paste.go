package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const PASTE = "paste"

func init() {
	Index.Register(&ice.Context{Name: PASTE, Help: "粘贴板",
		Configs: map[string]*ice.Config{
			PASTE: {Name: "paste", Help: "粘贴板", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(PASTE) }},

			PASTE: {Name: "paste hash=auto auto 添加 导出 导入", Help: "粘贴板", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert text:textarea=hi", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, arg)
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_TEXT, m.Option("text"))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					text := m.Cmd(mdb.SELECT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_HASH, arg[0]).Append("text")
					m.Cmdy("web.wiki.spark", "inner", text)
					m.Cmdy("web.wiki.image", "qrcode", text)
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
