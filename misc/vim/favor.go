package vim

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,type,name,text,file,line,pwd",
			)},
		},
		Commands: map[string]*ice.Command{
			"/favor": {Name: "/favor", Help: "收藏", Action: map[string]*ice.Action{
				mdb.SELECT: {Name: "select", Help: "主题", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(FAVOR).Table(func(index int, value map[string]string, head []string) {
						m.Echo(value[kit.MDB_ZONE]).Echo("\n")
					})
				}},
				mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(FAVOR, mdb.INSERT, m.OptionSimple("zone,type,name,text,file,line,pwd"))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(FAVOR, m.Option(kit.MDB_ZONE)).Table(func(index int, value map[string]string, head []string) {
					m.Echo("%v\n", m.Option(kit.MDB_ZONE)).Echo("%v:%v:%v:(%v): %v\n",
						value[kit.MDB_FILE], value[kit.MDB_LINE], "1", value[kit.MDB_NAME], value[kit.MDB_TEXT])
				})
			}},
			FAVOR: {Name: "favor zone id auto create export import", Help: "收藏夹", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert zone=数据结构 name=hi text=hello file line", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.LIST, arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(FAVOR), "", mdb.ZONE, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(FAVOR), "", mdb.ZONE, m.OptionSimple(kit.MDB_ZONE))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(kit.MDB_ZONE, m.Conf(FAVOR, kit.META_FIELD))
					m.Cmdy(mdb.EXPORT, m.Prefix(FAVOR), "", mdb.ZONE)
					m.Conf(FAVOR, kit.MDB_HASH, "")
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(FAVOR), "", mdb.ZONE)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_ZONE:
						m.Cmdy(mdb.INPUTS, m.Prefix(FAVOR), "", mdb.HASH, arg)
					default:
						m.Cmdy(mdb.INPUTS, m.Prefix(FAVOR), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.LIST, arg)
					}
				}},
				code.INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
					m.ProcessCommand(code.INNER, kit.Format([]string{
						kit.Select("./", path.Dir(m.Option(kit.MDB_FILE))),
						path.Base(m.Option(kit.MDB_FILE)),
						m.Option(kit.MDB_LINE),
					}), arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) < 2, kit.Select(m.Conf(FAVOR, kit.META_FIELD), "time,zone,count", len(arg) == 0))
				if m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), "", mdb.ZONE, arg); len(arg) == 0 {
					m.Action(mdb.CREATE)
					m.PushAction(mdb.REMOVE)
				} else {
					m.PushAction(code.INNER)
				}
			}},
		},
	})
}
