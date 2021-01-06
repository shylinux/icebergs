package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _github_show(m *ice.Message, name, text string, arg ...string) {
}
func _github_search(m *ice.Message, kind, name, text string) {
	m.PushSearch("cmd", GITHUB, "type", "github", "name", name, "text",
		kit.MergeURL("https://github.com/search", "q", name))
}

const GITHUB = "github"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		GITHUB: {Name: "github [name] word", Help: "百度", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_github_search(m, arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_github_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}})
}
