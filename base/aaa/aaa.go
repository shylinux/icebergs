package aaa

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const (
	ErrNotAuth = "not auth: "
)
const (
	IP = "ip"
	UA = "ua"

	USERROLE = "userrole"
	USERNAME = "username"
	PASSWORD = "password"
	USERNICK = "usernick"
	USERZONE = "userzone"

	SESSID = "sessid"
)
const AAA = "aaa"

var Index = &ice.Context{Name: AAA, Help: "认证模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Rich(ROLE, nil, kit.Dict(kit.MDB_NAME, VOID, White, kit.Dict(), Black, kit.Dict()))
		m.Rich(ROLE, nil, kit.Dict(kit.MDB_NAME, TECH, Black, kit.Dict(), White, kit.Dict()))
		m.Load()
		m.Cmd(mdb.SEARCH, mdb.CREATE, USER, USER, AAA)
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save()
	}},
}}

func init() { ice.Index.Register(Index, nil, USER, SESS, ROLE, TOTP) }
