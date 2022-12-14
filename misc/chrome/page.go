package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
)

type page struct {
	daemon
	style
	field

	list string `name:"list zone auto" help:"网页" http:"/page"`
}

func (p page) Command(m *ice.Message, arg ...string) {
	m.Cmdy(p.style.Command, arg)
	m.Cmdy(p.field.Command, arg)
}
func (p page) Run(m *ice.Message, arg ...string) {
	m.Cmdy(p.field.Run, arg)
}
func (p page) List(m *ice.Message, arg ...string) {
	p.daemon.Inputs(m, mdb.ZONE)
}

func init() { ice.CodeCtxCmd(page{}) }
