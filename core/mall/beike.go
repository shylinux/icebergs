package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _beike_show(m *ice.Message, name, text string, arg ...string) {
}
func _beike_search(m *ice.Message, kind, name, text string) {
	if kit.Contains(kind, kit.MDB_FOREACH) || kit.Contains(kind, BEIKE) {
		m.PushSearchWeb(BEIKE, name)
	}
}

const BEIKE = "beike"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			BEIKE: {Name: "beike", Help: "贝壳", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			BEIKE: {Name: "beike [name] word", Help: "百度", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_beike_search(m, arg[0], arg[1], arg[2])
				}},
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(mdb.INSERT, m.Prefix(BEIKE), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_beike_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		}})
}
