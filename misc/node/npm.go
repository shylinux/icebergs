package node

import "shylinux.com/x/ice"

type npm struct {
	ice.Code
	list string `name:"list auto" help:"打包构建"`
}

func (s npm) List(m *ice.Message) {

}

func init() { ice.CodeCtxCmd(npm{}) }
