package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
)

const NPM = "npm"

type npm struct {
	ice.Code
	list string `name:"list path auto" help:"依赖库"`
}

func (s npm) Init(m *ice.Message) {
	cli.IsAlpine(m.Message, NPM)
}
func (s npm) List(m *ice.Message) {
	m.Cmdy(nfs.DIR, ice.USR_MODULES)
}

func init() { ice.CodeCtxCmd(npm{}) }
