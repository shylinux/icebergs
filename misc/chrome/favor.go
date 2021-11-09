package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		FAVOR: {Name: "favor zone id auto", Help: "收藏夹", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=官网文档 type name=hi text=hello", Help: "添加"},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
			} else {
				m.PushAction(code.INNER)
			}
		}},
	}})
}
