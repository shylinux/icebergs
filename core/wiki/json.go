package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const JSON = "json"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			JSON: {Name: JSON, Help: "数据结构", Value: kit.Data(
				kit.MDB_PATH, ice.USR_LOCAL_EXPORT, kit.MDB_REGEXP, ".*\\.json",
			)},
		},
		Commands: map[string]*ice.Command{
			JSON: {Name: "json path auto", Help: "数据结构", Meta: kit.Dict(ice.Display("/plugin/local/wiki/json.js")), Action: map[string]*ice.Action{
				nfs.SAVE: {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_wiki_save(m, JSON, arg[0], arg[1])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if !_wiki_list(m, JSON, kit.Select("./", arg, 0)) {
					m.Cmdy(nfs.CAT, arg[0])
				}
			}},
		},
	})
}
