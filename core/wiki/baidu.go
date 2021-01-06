package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _baidu_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, BAIDU, name, text, arg...)
	// m.Cmdy(mdb.RENDER, web.RENDER.Frame, kit.Format("https://baidu.com/s?wd=%s", text))
}
func _baidu_search(m *ice.Message, kind, name, text string) {
	if kit.Contains(kind, "*") || kit.Contains(kind, "baidu") {
		m.PushSearchWeb(BAIDU, name)
	}
}

const BAIDU = "baidu"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			BAIDU: {Name: "baidu", Help: "百度", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			BAIDU: {Name: "baidu [name] word", Help: "百度", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_baidu_search(m, arg[0], arg[1], arg[2])
				}},
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(mdb.INSERT, m.Prefix(BAIDU), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_baidu_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		}})
}
