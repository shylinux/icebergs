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
	if kit.Contains(kind, kit.MDB_FOREACH) || kit.Contains(kind, MUSIC) {
		m.PushSearchWeb(MUSIC, name)
	}
}

const MUSIC = "music"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			MUSIC: {Name: "music", Help: "音乐", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			MUSIC: {Name: "music [name] url", Help: "视频", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_music_search(m, arg[0], arg[1], arg[2])
				}},
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(mdb.INSERT, m.Prefix(MUSIC), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_music_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		}})
}
