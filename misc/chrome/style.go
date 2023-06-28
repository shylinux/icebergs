package chrome

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

const (
	SELECTOR = "selector"
	PROPERTY = "property"
)

type style struct {
	ice.Zone
	daemon
	short  string `data:"domain"`
	field  string `data:"time,id,selector,property"`
	insert string `name:"insert domain=golang.google.cn selector=. property:textarea"`
	list   string `name:"style domain id auto insert" help:"样式"`
}

func (s style) Inputs(m *ice.Message, arg ...string) {
	s.daemon.Inputs(m, arg...)
}
func (s style) Command(m *ice.Message, arg ...string) {
	s.Zone.List(m, m.Option(web.DOMAIN)).Table(func(value ice.Maps) {
		s.send(m, "1", m.Option(TID), m.CommandKey(), value[SELECTOR], value[PROPERTY])
	})
	style := m.Cmdx(nfs.CAT, path.Join("src/website", m.Option(web.DOMAIN), "dark.css"))
	if style != "" {
		s.send(m, "1", m.Option(TID), m.CommandKey(), "style", style)
	}
}
func (s style) List(m *ice.Message, arg ...string) {
	s.Zone.List(m, arg...)
}
func init() { ice.CodeCtxCmd(style{}) }
