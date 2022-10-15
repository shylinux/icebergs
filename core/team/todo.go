package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/cli"
)

const TODO = "todo"

func init() {
	Index.MergeCommands(ice.Commands{
		TODO: {Name: "todo hash list create export import", Help: "待办", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashInputs(m, arg)
				// m.Cmdy(TASK, mdb.INPUTS, arg)
			}},
			mdb.CREATE: {Name: "create zone name text", Help: "创建"},
			cli.START: {Name: "start type=once,step,week", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(TASK, mdb.INSERT, m.OptionSimple("zone,type,name,text"), BEGIN_TIME, m.Time())
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,zone,name,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(cli.START, mdb.REMOVE)
			web.PushPodCmd(m, m.CommandKey(), arg...)
			ctx.DisplayTableCard(m)
		}},
	})
}
