package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/nfs"
)

type npm struct {
	ice.Code
	list string `name:"list auto"`
}

func (s npm) List(m *ice.Message) {
	m.Cmdy(nfs.DIR, ice.USR_MODULES)
}

func init() { ice.CodeCtxCmd(npm{}) }
