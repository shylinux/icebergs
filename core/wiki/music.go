package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _music_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, MUSIC, name, text, arg...)
}
func _music_search(m *ice.Message, kind, name, text string) {
	m.PushSearch("cmd", MUSIC, "type", "163", "name", name, "text",
		kit.MergeURL("https://music.163.com/#/search/m/", "s", name))
}

const MUSIC = "music"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		MUSIC: {Name: "music [name] url", Help: "视频", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_music_search(m, arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			arg = _name(m, arg)
			_music_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}})
}
