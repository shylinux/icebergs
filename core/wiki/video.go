package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _video_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, VIDEO, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(VIDEO, kit.Keym(kit.MDB_TEMPLATE)))
}
func _video_search(m *ice.Message, kind, name, text string) {
	if kit.Contains(kind, "*") || kit.Contains(kind, VIDEO) {
		m.PushSearchWeb(VIDEO, name)
	}
}

const VIDEO = "video"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			VIDEO: {Name: "video", Help: "视频", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_TEXT, kit.MDB_TEMPLATE, video,
			)},
		},
		Commands: map[string]*ice.Command{
			VIDEO: {Name: "video [name] url", Help: "视频", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_video_search(m, arg[0], arg[1], arg[2])
				}},
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(mdb.INSERT, m.Prefix(VIDEO), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_video_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		}})
}
