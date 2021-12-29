package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	SELECTOR = "selector"
	PROPERTY = "property"
)
const CHANGE = "change"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CHANGE: {Name: "change", Help: "编辑", Value: kit.Data(kit.MDB_SHORT, PROPERTY)},
	}, Commands: map[string]*ice.Command{
		CHANGE: {Name: "change wid tid selector:text@key property:textarea@key auto export import", Help: "编辑", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case SELECTOR:
					m.Push(arg[0], "video")
				default:
					m.Cmdy(mdb.INPUTS, m.PrefixKey(), "", mdb.HASH, arg)
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 || arg[2] == "" {
				m.Cmdy(web.SPACE, CHROME, CHROME, kit.Slice(arg, 0, 2))
				return
			}
			if len(arg) > 3 {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, SELECTOR, arg[2], PROPERTY, arg[3])
			}

			msg := m.Cmd(web.SPACE, CHROME, CHROME, kit.Slice(arg, 0, 2), CHANGE, kit.Slice(arg, 2))
			msg.Table(func(index int, value map[string]string, head []string) {
				m.Push(kit.MDB_TEXT, kit.ReplaceAll(value[kit.MDB_TEXT], "<", "&lt;", ">", "&gt;"))
			})
		}},
	}})
}
