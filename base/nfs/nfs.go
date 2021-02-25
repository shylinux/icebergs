package nfs

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "nfs", Help: "存储模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Load()
		m.Cmd(mdb.RENDER, mdb.CREATE, CAT)
		m.Cmd(mdb.SEARCH, mdb.CREATE, DIR)
		m.Cmd(mdb.RENDER, mdb.CREATE, DIR)

		m.Richs(TAIL, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			value, _ = kit.GetMeta(value), m.Option(kit.MDB_HASH, key)
			m.Cmd(TAIL, mdb.CREATE, kit.MDB_FILE, kit.Format(value[kit.MDB_FILE]), kit.MDB_NAME, kit.Format(value[kit.MDB_NAME]))
		})
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save()
	}},
}}

func init() { ice.Index.Register(Index, nil, CAT, DIR, TAIL, TRASH, SAVE, PUSH, COPY, LINK, DEFS) }
