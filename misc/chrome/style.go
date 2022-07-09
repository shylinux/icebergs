package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/tcp"
)

type style struct {
	ice.Zone
	operate

	short  string `data:"zone"`
	field  string `data:"time,id,selector,property"`
	insert string `name:"insert zone=golang.google.cn selector=. property:textarea" help:"添加"`
	list   string `name:"style zone id auto insert" help:"样式"`
}

func (s style) Inputs(m *ice.Message, arg ...string) {
	s.operate.Inputs(m, arg...)
}
func (s style) Command(m *ice.Message, arg ...string) {
	s.Zone.List(m, m.Option(tcp.HOST)).Table(func(index int, value ice.Maps, head []string) {
		s.send(m, "1", m.Option(TID), m.CommandKey(), value[SELECTOR], value[PROPERTY])
	})
}
func (s style) List(m *ice.Message, arg ...string) {
	s.Zone.List(m, arg...)
}

func init() { ice.CodeCtxCmd(style{}) }
