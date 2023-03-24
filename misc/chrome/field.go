package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

type field struct {
	ice.Zone
	daemon
	short  string `data:"domain"`
	field  string `data:"time,id,index,args,style,left,top,right,bottom,selection"`
	insert string `name:"insert domain=golang.google.cn index=cli.system args=pwd"`
	list   string `name:"list domain id auto insert" help:"插件"`
}

func (s field) Inputs(m *ice.Message, arg ...string) {
	s.daemon.Inputs(m, arg...)
}
func (s field) Command(m *ice.Message, arg ...string) {
	s.Zone.List(m.Spawn(), kit.Simple(m.Option(web.DOMAIN), arg)...).Table(func(index int, value ice.Maps, head []string) {
		if len(arg) == 0 {
			m.Options(ice.MSG_OPTS, head, kit.Simple(value))
			s.send(m, "1", m.Option(TID), m.CommandKey(), value[mdb.ID], value[ctx.ARGS])
		} else {
			m.OptionFields("")
			m.Cmdy(ctx.COMMAND, value[mdb.INDEX])
		}
	})
}
func (s field) Run(m *ice.Message, arg ...string) {
	s.Zone.List(m.Spawn(), m.Option(web.DOMAIN), arg[0]).Table(func(value ice.Maps) {
		m.Cmdy(value[mdb.INDEX], arg[1:])
	})
}
func (s field) List(m *ice.Message, arg ...string) {
	s.Zone.List(m, arg...)
}
func init() { ice.CodeCtxCmd(field{}) }
