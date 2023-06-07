package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

type scss struct {
	ice.Code
	ice.Lang
}

func (s scss) Init(m *ice.Message) {
	s.Lang.Init(m, code.PREPARE, ice.Map{
		code.KEYWORD:  kit.List("h1"),
		code.FUNCTION: kit.List(),
	}, "include", kit.List(nfs.CSS), "split.operator", "{[(.,:</>#)]}")
}
func (s scss) List(m *ice.Message) { m.Cmdy(nfs.DIR, nfs.USR) }

func init() { ice.CodeCtxCmd(scss{}) }
