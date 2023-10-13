package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const TODO = "todo"

func init() {
	Index.MergeCommands(ice.Commands{
		TODO: {Help: "待办", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { mdb.HashInputs(m, arg).Cmdy(TASK, mdb.INPUTS, arg) }},
			cli.START: {Name: "start type=once,step,week", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(TASK, mdb.INSERT, m.OptionSimple("zone,type,name,text"))
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
		}, mdb.ExportHashAction(mdb.FIELD, "time,hash,zone,name,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(cli.START, mdb.REMOVE)
			web.PushPodCmd(m, "", arg...)
			ctx.DisplayTableCard(m)
		}},
	})
}
