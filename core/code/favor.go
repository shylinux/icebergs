package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _sub_key(m *ice.Message, topic string) string {
	return kit.Keys(kit.MDB_HASH, kit.Hashs(topic))
}

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TOPIC)},
		},
		Commands: map[string]*ice.Command{
			FAVOR: {Name: "favor topic id auto insert export import", Help: "收藏夹", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create topic", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, FAVOR, "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert topic=数据结构 type=go name=hi text=hello path file line", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, FAVOR, "", mdb.HASH, kit.MDB_TOPIC, arg[1])
					m.Cmdy(mdb.INSERT, FAVOR, _sub_key(m, m.Option(kit.MDB_TOPIC)), mdb.LIST, arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, FAVOR, _sub_key(m, m.Option(kit.MDB_TOPIC)), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, FAVOR, "", mdb.HASH, kit.MDB_TOPIC, m.Option(kit.MDB_TOPIC))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, FAVOR, "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, FAVOR, "", mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_TOPIC:
						m.Cmdy(mdb.INPUTS, FAVOR, "", mdb.HASH, arg)
					default:
						m.Cmdy(mdb.INPUTS, FAVOR, _sub_key(m, m.Option(kit.MDB_TOPIC)), mdb.LIST, arg)
					}
				}},

				INNER: {Name: "inner", Help: "源代码", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) > 0 && arg[0] == mdb.RENDER {
						m.Cmdy(INNER, arg[1:])
						return
					}

					m.ShowPlugin("", m.Prefix(), INNER, INNER, mdb.RENDER)
					m.Push(kit.SSH_ARG, kit.Format([]string{m.Option(kit.MDB_PATH), m.Option(kit.MDB_FILE), m.Option(kit.MDB_LINE)}))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,count,topic", kit.Select("time,id,type,name,text,path,file,line", mdb.DETAIL, len(arg) > 1), len(arg) > 0))
				m.Cmdy(mdb.SELECT, FAVOR, "", mdb.ZONE, arg)
				m.PushAction(kit.Select(mdb.REMOVE, INNER, len(arg) > 0))
			}},
		},
	})
}
