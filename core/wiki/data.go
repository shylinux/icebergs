package wiki

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

func _data_show(m *ice.Message, name string, arg ...string) {
	m.Cmdy(nfs.CAT, name)
	m.CSV(m.Result())
}

const DATA = "data"
const (
	DataPlugin = "/plugin/local/wiki/data.js"
)

func init() {
	Index.Merge(&ice.Context{Name: "data", Help: "数据表格",
		Configs: map[string]*ice.Config{
			DATA: {Name: "data", Help: "数据表格", Value: kit.Data(
				"path", "usr/export", "regs", ".*\\.csv",
			)},
		},
		Commands: map[string]*ice.Command{
			DATA: {Name: "data path=auto auto", Help: "数据表格", Meta: kit.Dict(mdb.PLUGIN, DataPlugin), Action: map[string]*ice.Action{
				nfs.SAVE: {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_wiki_save(m, DATA, arg[0], arg[1])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if !_wiki_list(m, DATA, kit.Select("./", arg, 0)) {
					_data_show(m, path.Join(m.Conf(DATA, "meta.path"), arg[0]))
				}
			}},
		},
	}, nil)
}
