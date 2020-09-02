package aaa

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const (
	ErrNotAuth = "not auth: "
)
const (
	USERZONE = "userzone"
	USERNICK = "usernick"
	USERNAME = "username"
	PASSWORD = "password"
	USERROLE = "userrole"
	USERNODE = "usernode"
	HOSTPORT = "hostport"

	SESSID = "sessid"
)

var Index = &ice.Context{Name: "aaa", Help: "认证模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Rich(ROLE, nil, kit.Dict(kit.MDB_NAME, VOID, White, kit.Dict(), Black, kit.Dict()))
		m.Rich(ROLE, nil, kit.Dict(kit.MDB_NAME, TECH, Black, kit.Dict(), White, kit.Dict()))
		m.Load()
		m.Cmd("mdb.search", "create", "user", "user", "aaa")
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save(USER, SESS, ROLE)
	}},
}}

func init() { ice.Index.Register(Index, nil, USER, SESS, ROLE) }
