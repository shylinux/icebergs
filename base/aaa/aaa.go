package aaa

import (
	"github.com/shylinux/icebergs"
	// "github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"
)

const (
	ROLE = "role"
	USER = "user"
	SESS = "sess"
)
const (
	USERROLE = "userrole"
	USERNAME = "username"
	PASSWORD = "password"
	USERNODE = "usernode"
	USERNICK = "usernick"

	SESSID = "sessid"
)

var Index = &ice.Context{Name: "aaa", Help: "认证模块", Commands: map[string]*ice.Command{
	ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Rich(ROLE, nil, kit.Dict(kit.MDB_NAME, TECH, Black, kit.Dict(), White, kit.Dict()))
		m.Rich(ROLE, nil, kit.Dict(kit.MDB_NAME, VOID, White, kit.Dict(), Black, kit.Dict()))
		m.Load()
	}},
	ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save(ROLE, USER, SESS)
	}},
}}

func init() { ice.Index.Register(Index, nil, ROLE, USER, SESS) }
