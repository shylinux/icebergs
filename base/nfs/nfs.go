package nfs

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

var Index = &ice.Context{Name: "nfs", Help: "存储模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Load()
		m.Cmd(mdb.RENDER, mdb.CREATE, CAT)
		m.Cmd(mdb.SEARCH, mdb.CREATE, DIR)
		m.Cmd(mdb.RENDER, mdb.CREATE, DIR)
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save()
	}},
}}

func init() { ice.Index.Register(Index, nil, CAT, DIR, TAIL, TRASH, SAVE, PUSH, COPY, LINK, DEFS) }
