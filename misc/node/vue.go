package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/nfs"
)

type vue struct {
	ice.Code
	list string `name:"list path auto" help:"vue"`
}

func (s vue) List(m *ice.Message) {
	m.Cmdy(nfs.DIR, nfs.USR)
}

func init() { ice.CodeCtxCmd(vue{}) }
