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
			PASTE: {Name: "paste hash auto create@paste", Help: "粘贴板", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=text name=hi data:textarea=hi", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_trans(arg, map[string]string{"data": "text"})
					m.Conf(PASTE, kit.Keys(m.Option(ice.MSG_DOMAIN), kit.MDB_META, kit.MDB_SHORT), kit.MDB_TEXT)
					m.Cmdy(mdb.INSERT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, m.Prefix(PASTE), m.Option(ice.MSG_DOMAIN), mdb.HASH, arg)
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
				m.PushAction(mdb.REMOVE)
			}},
		},
	}, nil)
}
