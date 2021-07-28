package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

const DATA = "data"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DATA: {Name: DATA, Help: "数据表格", Value: kit.Data(
				kit.MDB_PATH, ice.USR_LOCAL_EXPORT, kit.MDB_REGEXP, ".*\\.csv",
			)},
		},
		Commands: map[string]*ice.Command{
			DATA: {Name: "data path auto", Help: "数据表格", Meta: kit.Dict(
				ice.Display("/plugin/local/wiki/data.js"),
			), Action: map[string]*ice.Action{
				nfs.SAVE: {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_wiki_save(m, DATA, arg[0], arg[1])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if !_wiki_list(m, DATA, kit.Select("./", arg, 0)) {
					m.CSV(m.Cmd(nfs.CAT, arg[0]).Result())
				}
			}},
		},
	})
}
