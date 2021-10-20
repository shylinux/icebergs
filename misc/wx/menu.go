package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const MENU = "menu"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		MENU: {Name: MENU, Help: "菜单", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,title,refer,image",
		)},
	}, Commands: map[string]*ice.Command{
		MENU: {Name: "menu zone id auto insert", Help: "菜单", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=home title=hi refer=hello image=", Help: "添加"},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(mdb.REMOVE)
			}
		}},
	}})
}
