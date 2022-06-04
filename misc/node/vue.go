package node

import "shylinux.com/x/ice"

type vue struct {
	ice.Code
	list string `name:"list auto" help:"服务框架"`
}

func (s vue) List(m *ice.Message) {

}

func init() { ice.CodeCtxCmd(vue{}) }
