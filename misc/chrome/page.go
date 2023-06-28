package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/web"
)

type page struct {
	daemon
	style
	field
	list string `name:"list domain auto" help:"网页" http:""`
}

func (s page) Command(m *ice.Message, arg ...string) {
	if m.Cmdy(s.field, s.field.Command, arg); len(arg) == 0 {
		m.Cmd(s.style, s.style.Command, arg)
	}
}
func (s page) Run(m *ice.Message, arg ...string) {
	m.Cmdy(s.field, s.Run, arg)
}
func (s page) List(m *ice.Message, arg ...string) {
	s.daemon.Inputs(m, web.DOMAIN)
}
func init() { ice.CodeCtxCmd(page{}) }
