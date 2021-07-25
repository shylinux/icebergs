package wx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const MENU = "menu"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			MENU: {Name: MENU, Help: "菜单", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,title,refer,image",
			)},
		},
		Commands: map[string]*ice.Command{
			MENU: {Name: "menu zone id auto create", Help: "菜单", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(MENU), "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert zone=home title=hi refer=hello image=", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(MENU), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
					m.Cmdy(mdb.INSERT, m.Prefix(MENU), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.LIST, arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(MENU), "", mdb.ZONE, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(MENU), "", mdb.ZONE, m.OptionSimple(kit.MDB_ZONE))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,zone,count", m.Conf(MENU, kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, m.Prefix(MENU), "", mdb.ZONE, arg); len(arg) == 0 {
					m.PushAction(mdb.REMOVE)
				}
			}},
		}})
}
