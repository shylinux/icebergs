package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _csdn_show(m *ice.Message, name, text string, arg ...string) {
}
func _csdn_search(m *ice.Message, kind, name, text string) {
	if kit.Contains(kind, "*") || kit.Contains(kind, "csdn") {
		m.PushSearchWeb(CSDN, name)
	}
}

const CSDN = "csdn"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CSDN: {Name: "csdn", Help: "博客", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		}, Commands: map[string]*ice.Command{
			CSDN: {Name: "csdn [name] word", Help: "博客", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_csdn_search(m, arg[0], arg[1], arg[2])
				}},
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(mdb.INSERT, m.Prefix(CSDN), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_csdn_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		}})
}
