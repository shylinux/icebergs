package vim

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(
				kit.MDB_SHORT, "topic", kit.MDB_FIELD, "time,id,type,name,text,file,line",
			)},
		},
		Commands: map[string]*ice.Command{
			FAVOR: {Name: "favor topic=auto id=auto auto 创建 导出 导入", Help: "收藏夹", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create topic", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert topic=数据结构 name=hi text=hello", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(m.Prefix(FAVOR), "", m.Option("topic"), func(key string, value map[string]interface{}) {
						m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), kit.Keys(kit.MDB_HASH, key), mdb.LIST, arg)
					})
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(m.Prefix(FAVOR), "", m.Option("topic"), func(key string, value map[string]interface{}) {
						m.Cmdy(mdb.MODIFY, m.Prefix(FAVOR), kit.Keys(kit.MDB_HASH, key), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID), arg)
					})
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(FAVOR), "", mdb.HASH, "topic", m.Option("topic"))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(FAVOR), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(FAVOR), "", mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "topic":
						m.Option(mdb.FIELDS, "time,count,topic")
						m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), "", mdb.HASH)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Option(mdb.FIELDS, kit.Select(m.Conf(m.Prefix(FAVOR), kit.META_FIELD), mdb.DETAIL, len(arg) > 1))
					m.Richs(m.Prefix(FAVOR), "", arg[0], func(key string, value map[string]interface{}) {
						m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), kit.Keys(kit.MDB_HASH, key), mdb.LIST, kit.MDB_ID, arg[1:])
					})
					return
				}
				m.Option(mdb.FIELDS, "time,count,topic")
				m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), "", mdb.HASH)
				m.PushAction("删除")
			}},

			"/favor": {Name: "/favor", Help: "收藏", Action: map[string]*ice.Action{
				"select": {Name: "select", Help: "主题", Hand: func(m *ice.Message, arg ...string) {
					list := []string{}
					m.Cmd(m.Prefix(FAVOR)).Table(func(index int, value map[string]string, head []string) {
						list = append(list, value["topic"])
					})
					m.Render(ice.RENDER_RESULT, strings.Join(list, "\n"))
				}},
				"insert": {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(m.Prefix(FAVOR), mdb.INSERT, "topic", m.Option("tab"),
						"name", m.Option("note"), "text", m.Option("arg"), "file", m.Option("buf"), "line", m.Option("row"),
					)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Render(ice.RENDER_RESULT)
				m.Cmd(m.Prefix(FAVOR), m.Option("tab")).Table(func(index int, value map[string]string, head []string) {
					m.Echo("%v\n", m.Option("tab")).Echo("%v:%v:%v:(%v): %v\n",
						value["file"], value["line"], "1", value["name"], value["text"])
				})
			}},
		},
	}, nil)
}
