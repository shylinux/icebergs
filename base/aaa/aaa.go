package aaa

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const AAA = "aaa"

var Index = &ice.Context{Name: AAA, Help: "认证模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Rich(ROLE, nil, kit.Dict(kit.MDB_NAME, VOID, WHITE, kit.Dict(), BLACK, kit.Dict()))
		m.Rich(ROLE, nil, kit.Dict(kit.MDB_NAME, TECH, BLACK, kit.Dict(), WHITE, kit.Dict()))
		m.Load()
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save()
	}},
}}

func init() { ice.Index.Register(Index, nil, ROLE, SESS, TOTP, USER) }
