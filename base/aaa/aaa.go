package aaa

import (
	ice "shylinux.com/x/icebergs"
)

const (
	RSA = "rsa"
)
const AAA = "aaa"

var Index = &ice.Context{Name: AAA, Help: "认证模块"}

func init() { ice.Index.Register(Index, nil, ROLE, SESS, TOTP, USER, RSA) }
