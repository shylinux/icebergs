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

			PASTE: {Name: "paste text auto 粘贴:button", Help: "粘贴板", Meta: kit.Dict(
				"display", "/plugin/story/paste.js",
			), Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert [key value]...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_TEXT, m.Option("text"), arg[0], arg[1])
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_TEXT, m.Option("text"))
				}},
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_TEXT, arg[0])
				}

				m.Cmdy(mdb.SELECT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				m.PushAction("复制", "删除")
				m.Sort("time", "time_r")
			}},
		},
	}, nil)
}
