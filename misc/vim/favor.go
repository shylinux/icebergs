package vim

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_TOPIC, kit.MDB_FIELD, "time,id,type,name,text,file,line",
			)},
		},
		Commands: map[string]*ice.Command{
			FAVOR: {Name: "favor topic=auto id=auto auto create export import", Help: "收藏夹", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create topic", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert topic=数据结构 name=hi text=hello", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, kit.MDB_TOPIC, m.Option(kit.MDB_TOPIC))
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), kit.SubKey(m.Option(kit.MDB_TOPIC)), mdb.LIST, arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(FAVOR), kit.SubKey(m.Option(kit.MDB_TOPIC)), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(FAVOR), "", mdb.HASH, kit.MDB_TOPIC, m.Option(kit.MDB_TOPIC))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(FAVOR), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(FAVOR), "", mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_TOPIC:
						m.Cmdy(mdb.INPUTS, m.Prefix(FAVOR), "", mdb.HASH, arg)
					default:
						m.Cmdy(mdb.INPUTS, m.Prefix(FAVOR), kit.SubKey(m.Option(kit.MDB_TOPIC)), mdb.LIST, arg)
					}
				}},
				code.INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) > 0 && arg[0] == mdb.RENDER {
						m.Cmdy(code.INNER, arg[1:])
						return
					}

					m.PushPlugin(code.INNER, code.INNER, mdb.RENDER)
					m.Push(kit.SSH_ARG, kit.Format([]string{kit.Select("./", m.Option(kit.MDB_PATH)), m.Option(kit.MDB_FILE), m.Option(kit.MDB_LINE)}))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,count,topic")
					m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), "", mdb.HASH)
					m.PushAction(mdb.REMOVE)
					return
				}

				m.Option(mdb.FIELDS, kit.Select(m.Conf(m.Prefix(FAVOR), kit.META_FIELD), mdb.DETAIL, len(arg) > 1))
				m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), kit.SubKey(arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				m.PushAction(code.INNER)
			}},

			"/favor": {Name: "/favor", Help: "收藏", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(m.Prefix(FAVOR), mdb.INSERT, kit.MDB_TOPIC, m.Option(TAB),
						kit.MDB_NAME, m.Option("note"), kit.MDB_TEXT, m.Option(ARG), kit.MDB_FILE, m.Option(BUF), kit.MDB_LINE, m.Option(ROW),
					)
				}},
				mdb.SELECT: {Name: "select", Help: "主题", Hand: func(m *ice.Message, arg ...string) {
					list := []string{}
					m.Cmd(m.Prefix(FAVOR)).Table(func(index int, value map[string]string, head []string) {
						list = append(list, value[kit.MDB_TOPIC])
					})
					m.Echo(strings.Join(list, "\n"))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(m.Prefix(FAVOR), m.Option(TAB)).Table(func(index int, value map[string]string, head []string) {
					m.Echo("%v\n", m.Option(TAB)).Echo("%v:%v:%v:(%v): %v\n",
						value[kit.MDB_FILE], value[kit.MDB_LINE], "1", value[kit.MDB_NAME], value[kit.MDB_TEXT])
				})
			}},
		},
	})
}
