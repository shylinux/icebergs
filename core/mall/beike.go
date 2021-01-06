package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net/url"
)

func _beike_show(m *ice.Message, name, text string, arg ...string) {
}
func _beike_search(m *ice.Message, kind, name, text string) {
	m.PushSearch("cmd", BEIKE, "type", kind, "name", name, "text",
		kit.MergeURL("https://ke.com/ershoufang/rs"+url.QueryEscape(name)))
}

const BEIKE = "beike"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		BEIKE: {Name: "beike [name] word", Help: "百度", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_beike_search(m, arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_beike_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}})
}
