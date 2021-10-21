package ctx

import (
	ice "shylinux.com/x/icebergs"
)

const CTX = "ctx"

var Index = &ice.Context{Name: CTX, Help: "标准模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
	}},
}}

func init() { ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG, MESSAGE) }
