package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _taobao_show(m *ice.Message, name, text string, arg ...string) {
}
func _taobao_search(m *ice.Message, kind, name, text string) {
	m.PushSearch("cmd", TAOBAO, "type", kind, "name", name, "text", kit.MergeURL("https://s.taobao.com/search", "q", name))
}

const TAOBAO = "taobao"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TAOBAO: {Name: "taobao [name] word", Help: "百度", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_taobao_search(m, arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_taobao_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}})
}
