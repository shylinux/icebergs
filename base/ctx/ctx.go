package ctx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const CTX = "ctx"

var Index = &ice.Context{Name: CTX, Help: "标准模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Cmd(mdb.SEARCH, mdb.CREATE, COMMAND, m.Prefix(COMMAND))
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
	}},
}}

func init() { ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG) }
