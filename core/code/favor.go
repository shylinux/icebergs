package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,type,name,text,path,file,line",
			)},
		},
		Commands: map[string]*ice.Command{
			FAVOR: {Name: "favor zone id auto insert export import", Help: "收藏夹", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert zone=数据结构 type=go name=hi text=hello path file line", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, kit.MDB_ZONE, arg[1])
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.ZONE, m.Option(kit.MDB_ZONE), arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(FAVOR), "", mdb.ZONE, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(FAVOR), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(kit.MDB_ZONE, m.Conf(FAVOR, kit.META_FIELD))
					m.Cmdy(mdb.EXPORT, m.Prefix(FAVOR), "", mdb.ZONE)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(kit.MDB_ZONE)
					m.Cmdy(mdb.IMPORT, m.Prefix(FAVOR), "", mdb.ZONE)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_ZONE:
						m.Cmdy(mdb.INPUTS, m.Prefix(FAVOR), "", mdb.HASH, arg)
					default:
						m.Cmdy(mdb.INPUTS, m.Prefix(FAVOR), "", mdb.ZONE, m.Option(kit.MDB_ZONE), arg)
					}
				}},
				INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
					m.ProcessCommand(INNER, m.OptionSplit("path,file,line"), arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,zone,count", m.Conf(FAVOR, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), "", mdb.ZONE, arg)
				m.PushAction(kit.Select(mdb.REMOVE, INNER, len(arg) > 0))
			}},
		},
	})
}
