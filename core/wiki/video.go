package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net/url"
)

func _video_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, VIDEO, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(VIDEO, "meta.template"))
}
func _video_search(m *ice.Message, kind, name, text string) {
	m.PushSearch("cmd", VIDEO, "type", "v", "name", name, "text",
		kit.MergeURL("https://v.qq.com/x/search/", "q", name))

	m.PushSearch("cmd", VIDEO, "type", "iqiyi", "name", name, "text",
		kit.MergeURL("https://so.iqiyi.com/so/q_"+url.QueryEscape(name)))

	m.PushSearch("cmd", VIDEO, "type", "kuaishou", "name", name, "text",
		kit.MergeURL("https://video.kuaishou.com/search", "searchKey", name))

	m.PushSearch("cmd", VIDEO, "type", "bilibili", "name", name, "text",
		kit.MergeURL("https://search.bilibili.com/all", "keyword", name))

}

const VIDEO = "video"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		VIDEO: {Name: "video [name] url", Help: "视频", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_video_search(m, arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			arg = _name(m, arg)
			_video_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}})
}
