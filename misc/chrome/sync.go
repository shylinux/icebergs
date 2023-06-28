package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
)

type sync struct {
	field  string `data:"time,id,type,name,link"`
	insert string `name:"insert type name link" http:"sync"`
	list   string `name:"list id auto" help:"同步流"`
}

func (s sync) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case mdb.ZONE:
		m.Cmdy(arg)
	default:
	}
}
func (s sync) Insert(m *ice.Message, arg ...string) {
}
func (s sync) List(m *ice.Message, arg ...string) {
}
func init() { ice.CodeCtxCmd(sync{}) }
