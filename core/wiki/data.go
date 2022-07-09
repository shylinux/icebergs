package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DATA = "data"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		DATA: {Name: DATA, Help: "数据表格", Value: kit.Data(nfs.PATH, ice.USR_LOCAL_EXPORT, lex.REGEXP, ".*\\.csv")},
	}, Commands: ice.Commands{
		DATA: {Name: "data path auto", Help: "数据表格", Meta: kit.Dict(ice.DisplayLocal("")), Actions: ice.Actions{
			nfs.SAVE: {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				_wiki_save(m, m.CommandKey(), arg[0], arg[1])
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if !_wiki_list(m, m.CommandKey(), kit.Select(nfs.PWD, arg, 0)) {
				m.CSV(m.Cmd(nfs.CAT, arg[0]).Result())
				m.StatusTimeCount()
			}
		}},
	}})
}
