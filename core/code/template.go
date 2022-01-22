package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
)

func init() {
	const TEMPLATE = "template"
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TEMPLATE: {Name: "template name auto create", Help: "模板", Action: ice.MergeAction(
			map[string]*ice.Action{
				mdb.CREATE: {Name: "create type name text args", Help: "创建"},
				nfs.DEFS: {Name: "defs file", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Option(mdb.TEXT))
				}},
			}, mdb.HashAction(mdb.SHORT, "name", mdb.FIELD, "time,type,name,text,args"),
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.PushAction(nfs.DEFS, mdb.REMOVE)
		}}},
	})
}
