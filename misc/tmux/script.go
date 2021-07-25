package tmux

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SCRIPT: {Name: SCRIPT, Help: "脚本", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			SCRIPT: {Name: "script name auto create export import", Help: "脚本", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type=shell,tmux,vim name=hi text:textarea=pwd", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(SCRIPT), "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(SCRIPT), "", mdb.HASH, m.OptionSimple(kit.MDB_NAME), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(SCRIPT), "", mdb.HASH, m.OptionSimple(kit.MDB_NAME))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(SCRIPT), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(SCRIPT), "", mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(SCRIPT, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(SCRIPT), "", mdb.HASH, kit.MDB_NAME, arg)
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
