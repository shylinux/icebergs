package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _github_show(m *ice.Message, name, text string, arg ...string) {
}
func _github_search(m *ice.Message, kind, name, text string) {
	if kit.Contains(kind, kit.MDB_FOREACH) || kit.Contains(kind, GITHUB) {
		m.PushSearchWeb(GITHUB, name)
	}
}

const GITHUB = "github"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			GITHUB: {Name: "github", Help: "仓库", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			GITHUB: {Name: "github [name] word", Help: "百度", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(mdb.INSERT, m.Prefix(GITHUB), "", mdb.HASH, arg)
				}},
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_github_search(m, arg[0], arg[1], arg[2])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_github_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		}})
}
