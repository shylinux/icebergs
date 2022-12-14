package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

type sync struct {
	ice.Lists
	favor `name:"favor zone=some type name link"`

	field  string `data:"time,id,type,name,link"`
	insert string `name:"insert type name link" http:"/sync"`
	list   string `name:"list id auto" help:"同步流"`
}

func (s sync) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case mdb.ZONE:
		m.Cmdy(s.favor.Inputs, arg)
	default:
		m.Cmdy(s.Lists.Inputs, arg)
	}
}
func (s sync) Insert(m *ice.Message, arg ...string) {
	s.Lists.Insert(m, arg...)
}
func (s sync) Favor(m *ice.Message, arg ...string) {
	m.Cmdy(s.favor.Insert, m.OptionSimple("zone,type,name,link"))
	web.ToastSuccess(m.Message)
}
func (s sync) List(m *ice.Message, arg ...string) {
	s.Lists.List(m, arg...)
	m.PushAction(s.Favor)
}

func init() { ice.CodeCtxCmd(sync{}) }
