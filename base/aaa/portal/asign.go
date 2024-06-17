package portal

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

type asign struct {
	ice.Hash
	export string `data:"true"`
	short  string `data:"role"`
	field  string `data:"time,role"`
	shorts string `data:"index"`
	fields string `data:"time,index,operate"`
	insert string `name:"insert index"`
	deploy string `name:"deploy" help:"部署"`
	list   string `name:"list role auto" help:"分配"`
	confer string `name:"confer username" help:"授权"`
}

func (s asign) Inputs(m *ice.Message, arg ...string) {
	if arg[0] == "operate" {
		m.Search(m.Option(ctx.INDEX), func(key string, cmd *ice.Command) {
			for sub, action := range cmd.Actions {
				if kit.HasPrefix(sub, "_", "/") {
					continue
				}
				m.Push(arg[0], sub)
				m.Push(mdb.NAME, action.Name)
				m.Push(mdb.HELP, action.Help)
			}
			m.Sort(arg[0])
			m.Option(ice.TABLE_CHECKBOX, ice.TRUE)
		})
	} else if arg[0] == aaa.USERNAME {
		m.Cmdy(aaa.USER).Cut(aaa.USERROLE, aaa.USERNAME, aaa.USERNICK)
	} else {
		s.Hash.Inputs(m, arg...)
	}
}
func (s asign) Modify(m *ice.Message, arg ...string) {
	if m.Option(ctx.INDEX) != "" {
		s.Update(m, arg...)
	} else {
		s.Modify(m, arg...)
	}
}
func (s asign) Deploy(m *ice.Message, arg ...string) {
	defer m.ToastProcess()()
	s.List(m.Spawn()).Table(func(val ice.Maps) {
		m.Cmd(aaa.ROLE, mdb.REMOVE, val[aaa.ROLE])
		m.Cmd(aaa.ROLE, mdb.CREATE, val[aaa.ROLE])
		s.List(m.Spawn(), val[aaa.ROLE]).Table(func(value ice.Maps) {
			m.Cmd(aaa.ROLE, aaa.WHITE, val[aaa.ROLE], value[ctx.INDEX])
			m.Cmd(aaa.ROLE, aaa.BLACK, val[aaa.ROLE], value[ctx.INDEX], ctx.ACTION)
			kit.For(kit.Split(value["operate"]), func(p string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, val[aaa.ROLE], value[ctx.INDEX], ctx.ACTION, p)
			})
		})
	})
}
func (s asign) List(m *ice.Message, arg ...string) *ice.Message {
	if len(arg) == 0 {
		s.Hash.List(m, arg...).Action(s.Create, s.Deploy).PushAction(s.Confer, s.Remove)
	} else {
		s.Hash.SubList(m, arg[0], arg[1:]...).Action(s.Insert, s.Deploy).PushAction(s.Delete)
	}
	return m
}
func (s asign) Confer(m *ice.Message, arg ...string) {
	m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERNAME, m.Option(aaa.USERNAME), aaa.USERROLE, m.Option(aaa.ROLE))
}

func init() { ice.Cmd("aaa.asign", asign{}) }
