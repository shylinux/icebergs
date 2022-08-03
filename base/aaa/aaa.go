package aaa

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

const (
	RSA = "rsa"
)
const AAA = "aaa"

var Index = &ice.Context{Name: AAA, Help: "认证模块"}

func init() { ice.Index.Register(Index, nil, ROLE, SESS, TOTP, USER, RSA) }

func Right(m *ice.Message, arg ...ice.Any) bool {
	return m.Option(ice.MSG_USERROLE) == ROOT || !m.Warn(m.Cmdx(ROLE, RIGHT, m.Option(ice.MSG_USERROLE), arg) != ice.OK,
		ice.ErrNotRight, kit.Join(kit.Simple(arg), ice.PT), USERROLE, m.Option(ice.MSG_USERROLE), logs.FileLineMeta(kit.FileLine(2, 3)))
}
