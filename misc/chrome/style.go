package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const STYLE = "style"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		STYLE: {Name: "style", Help: "样式", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,target,style",
		)},
	}, Commands: map[string]*ice.Command{
		STYLE: {Name: "style zone id auto insert", Help: "样式", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_ZONE:
					m.Cmdy(CHROME, mdb.INPUTS)
				default:
					m.Cmdy(mdb.INPUTS, m.PrefixKey(), "", mdb.ZONE, m.Option(m.Config(kit.MDB_SHORT)), arg)
				}
			}},
			mdb.INSERT: {Name: "insert zone=golang.google.cn target=. style:textarea", Help: "添加"},
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(STYLE, m.Option(tcp.HOST)).Table(func(index int, value map[string]string, head []string) {
					m.Cmdy(web.SPACE, CHROME, CHROME, "1", m.Option("tid"), STYLE, value["target"], value["style"])
				})
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(mdb.REMOVE)
			}
		}},
	}})
}
