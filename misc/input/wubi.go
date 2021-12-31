package input

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const WUBI = "wubi"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		WUBI: {Name: WUBI, Help: "输入法", Value: kit.Data(
			kit.MDB_STORE, path.Join(ice.USR_LOCAL_EXPORT, INPUT, WUBI), kit.MDB_FSIZE, "200000",
			kit.MDB_SHORT, "zone", nfs.REPOS, "wubi-dict",
			kit.MDB_LIMIT, "5000", kit.MDB_LEAST, "1000",
		)},
	}, Commands: map[string]*ice.Command{
		WUBI: {Name: "wubi method=word,line code auto", Help: "五笔", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=person text code weight", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_input_push(m, m.Option(ZONE), m.Option(TEXT), m.Option(CODE), m.Option(WEIGHT))
			}},
			mdb.EXPORT: {Name: "export file=usr/wubi-dict/person zone=person", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				_input_save(m, m.Option(FILE), m.Option(ZONE))
			}},
			mdb.IMPORT: {Name: "import file=usr/wubi-dict/wubi86 zone=wubi86", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				_input_load(m, m.Option(FILE), m.Option(ZONE))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_input_find(m, arg[0], arg[1], m.Option(ice.CACHE_LIMIT))
			m.StatusTime()
		}},
	}})
}
