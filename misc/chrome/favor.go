package chrome

import (
	"shylinux.com/x/ice"
)

type favor struct {
	ice.Zone
	short  string `data:"zone"`
	field  string `data:"time,id,type,name,link"`
	insert string `name:"insert zone*=官网文档 type name=hi link=*hello"`
	list   string `name:"list zone id auto insert" help:"收藏"`
}

func (f favor) List(m *ice.Message, arg ...string) {
	if f.Zone.List(m, arg...); len(arg) == 0 {
		m.Action(f.Export, f.Import)
	}
}
func init() { ice.CodeCtxCmd(favor{}) }
