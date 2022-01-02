package chrome

import (
	"shylinux.com/x/ice"
)

type favor struct {
	ice.Zone

	short  string `data:"zone"`
	field  string `data:"time,id,type,name,link"`
	insert string `name:"insert zone=官网文档 type name=hi link=hello" help:"添加"`
	list   string `name:"list zone id auto" help:"收载"`
}

func (f favor) Inputs(m *ice.Message, arg ...string) {
	f.Zone.Inputs(m, arg...)
}
func (f favor) Insert(m *ice.Message, arg ...string) {
	f.Zone.Insert(m, arg...)
}
func (f favor) List(m *ice.Message, arg ...string) {
	if f.Zone.List(m, arg...); len(arg) == 0 {
		m.Action(f.Insert, f.Export, f.Import)
	} else if len(arg) == 1 {
		m.Action(f.Insert)
	}
}
func init() { ice.CodeCtxCmd(favor{}) }
