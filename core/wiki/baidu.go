package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net/url"
	"strings"
)

func _baidu_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, BAIDU, name, text, arg...)
	// m.Cmdy(mdb.RENDER, web.RENDER.Frame, kit.Format("https://baidu.com/s?wd=%s", text))
}
func _baidu_search(m *ice.Message, kind, name, text string) {
	if !strings.Contains(kind, "*") && !strings.Contains(kind, "baidu") {
		return
	}

	m.PushSearch("cmd", BAIDU, "type", "web", "name", name, "text",
		kit.MergeURL("https://www.baidu.com/s", "wd", name))

	m.PushSearch("cmd", BAIDU, "type", "map", "name", name, "text",
		kit.MergeURL("https://map.baidu.com/search?querytype=s", "wd", name))

	m.PushSearch("cmd", BAIDU, "type", "zhihu", "name", name, "text",
		kit.MergeURL("https://www.zhihu.com/search?type=content", "q", name))

	m.PushSearch("cmd", BAIDU, "type", "weibo", "name", name, "text",
		kit.MergeURL("https://s.weibo.com/weibo/"+url.QueryEscape(name)))

	m.PushSearch("cmd", BAIDU, "type", "taotiao", "name", name, "text",
		kit.MergeURL("https://www.toutiao.com/search/", "keyword", name))
}

const BAIDU = "baidu"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		BAIDU: {Name: "baidu [name] word", Help: "百度", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_baidu_search(m, arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			arg = _name(m, arg)
			_baidu_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}})
}
