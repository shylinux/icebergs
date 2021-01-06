package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
	"strings"
)

func _csdn_show(m *ice.Message, name, text string, arg ...string) {
}
func _csdn_search(m *ice.Message, kind, name, text string) {
	if !strings.Contains(kind, "*") && !strings.Contains(kind, "csdn") {
		return
	}

	m.PushSearch("cmd", CSDN, "type", "csdn", "name", name, "text",
		kit.MergeURL("https://so.csdn.net/so/search/all", "q", name))

	m.PushSearch("cmd", CSDN, "type", "juejin", "name", name, "text",
		kit.MergeURL("https://juejin.cn/search?type=all", "query", name))

	m.PushSearch("cmd", CSDN, "type", "51cto", "name", name, "text",
		kit.MergeURL("http://so.51cto.com/?sort=time", "keywords", name))

	m.PushSearch("cmd", CSDN, "type", "oschina", "name", name, "text",
		kit.MergeURL("https://www.oschina.net/search?scope=project", "q", name))

	m.PushSearch("cmd", CSDN, "type", "imooc", "name", name, "text",
		kit.MergeURL("https://www.imooc.com/search/", "words", name))

	m.PushSearch("cmd", CSDN, "type", "segmentfault", "name", name, "text",
		kit.MergeURL("https://segmentfault.com/search", "q", name))

	m.PushSearch("cmd", CSDN, "type", "nowcoder", "name", name, "text",
		kit.MergeURL("https://www.nowcoder.com/search?type=all", "query", name))

	m.PushSearch("cmd", CSDN, "type", "leetcode-cn", "name", name, "text",
		kit.MergeURL("https://leetcode-cn.com/problemset/all/", "search", name))
}

const CSDN = "csdn"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		CSDN: {Name: "csdn [name] word", Help: "百度", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_csdn_search(m, arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_csdn_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}})
}
