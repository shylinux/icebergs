package aaa

import (
	ice "shylinux.com/x/icebergs"
)

const (
	RSA    = "rsa"
	SIGN   = "sign"
	VERIFY = "verify"
	BASE64 = "base64"
)
const AAA = "aaa"

var Index = &ice.Context{Name: AAA, Help: "认证模块"}

func init() { ice.Index.Register(Index, nil, APPLY, OFFER, EMAIL, USER, TOTP, SESS, ROLE, RSA) }
