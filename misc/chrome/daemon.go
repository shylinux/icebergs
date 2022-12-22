package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	WID = "wid"
	TID = "tid"
	URL = "url"
)

type daemon struct {
	list string `name:"list wid tid url auto" help:"后台"`
}

func (s daemon) send(m *ice.Message, arg ...ice.Any) *ice.Message {
	return m.Cmdy(web.SPACE, "chrome", "chrome", arg)
}
func (s daemon) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case web.DOMAIN:
		s.send(m.Spawn()).Tables(func(value ice.Maps) {
			s.send(m.Spawn(), value[WID]).Tables(func(value ice.Maps) {
				if value[URL] != "" {
					m.Push(arg[0], kit.ParseURL(value[URL]).Host)
				}
			})
		m.Debug("what %v", m.FormatsMeta())
		}).Sort(arg[0])
		m.Debug("what %v", m.FormatsMeta())
	case ctx.INDEX:
		ctx.CmdList(m.Message)
	}
}
func (s daemon) List(m *ice.Message, arg ...string) {
	if len(arg) < 3 || arg[0] == "" || arg[1] == "" {
		s.send(m, arg).StatusTimeCount()
	} else {
		s.send(m, arg[:2], "user.jumps", arg[2])
	}
}
func init() { ice.CodeCtxCmd(daemon{}) }
