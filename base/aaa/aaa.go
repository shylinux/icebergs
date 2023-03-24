package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const (
	RSA = "rsa"
)
const AAA = "aaa"

var Index = &ice.Context{Name: AAA, Help: "认证模块", Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		ice.Info.Load(m).Cmd(ROLE, ice.CTX_INIT).Cmd(ROLE, mdb.CREATE, TECH, VOID)
	}},
}}

func init() { ice.Index.Register(Index, nil, OFFER, EMAIL, USER, TOTP, SESS, ROLE, RSA) }
