package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

type field struct {
	ice.Zone
	daemon

	short  string `data:"zone"`
	field  string `data:"time,id,index,args,style,left,top,right,bottom,selection"`
	insert string `name:"insert zone=golang.google.cn index=cli.system args=pwd"`
	list   string `name:"list zone id auto insert" help:"插件"`
}

func (f field) Inputs(m *ice.Message, arg ...string) {
	f.daemon.Inputs(m, arg...)
}
func (f field) Command(m *ice.Message, arg ...string) {
	m.OptionFields("")
	f.Zone.List(m.Spawn(), kit.Simple(m.Option(tcp.HOST), arg)...).Table(func(index int, value ice.Maps, head []string) {
		if len(arg) == 0 {
			m.Option(ice.MSG_OPTS, head)
			m.Options(kit.Simple(value))
			f.send(m.Spawn(), "1", m.Option(TID), m.CommandKey(), value[mdb.ID], value[ctx.ARGS])
		} else {
			m.Cmdy(ctx.COMMAND, value[mdb.INDEX])
		}
	})
}
func (f field) Run(m *ice.Message, arg ...string) {
	f.Zone.List(m.Spawn(), m.Option(tcp.HOST), arg[0]).Tables(func(value ice.Maps) {
		m.Cmdy(value[mdb.INDEX], arg[1:])
	})
}
func (f field) List(m *ice.Message, arg ...string) {
	f.Zone.List(m, arg...)
}

func init() { ice.CodeCtxCmd(field{}) }
