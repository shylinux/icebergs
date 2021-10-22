package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const STYLE = "style"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		STYLE: {Name: "style", Help: "样式", Value: kit.Data(
			kit.MDB_SHORT, "zone", kit.MDB_FIELD, "time,id,target,style",
		)},
	}, Commands: map[string]*ice.Command{
		STYLE: {Name: "style zone id auto insert", Help: "样式", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=golang.google.cn target=. style:textarea", Help: "添加"},
			SYNC: {Name: "sync hostname", Help: "同步", Hand: func(m *ice.Message, arg ...string) {
				m.Fields(0, m.Conf(STYLE, kit.META_FIELD))
				m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.ZONE, m.Option("hostname")).Table(func(index int, value map[string]string, head []string) {
					m.Cmd(web.SPACE, CHROME, CHROME, "1", m.Option("tid"), STYLE, value["target"], value["style"])
				})
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(mdb.REMOVE)
			}
		}},
	}})
}
