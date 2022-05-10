package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
)

const TODO = "todo"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TODO: {Name: "todo hash list create export import", Help: "待办", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INPUTS, m.PrefixKey(), "", mdb.HASH, arg)
				m.Cmdy(TASK, mdb.INPUTS, arg)
			}},
			mdb.CREATE: {Name: "create zone name text", Help: "创建"},
			cli.START: {Name: "start type=once,step,week", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(TASK, mdb.INSERT, m.OptionSimple("zone,type,name,text"), BEGIN_TIME, m.Time())
				m.Cmd(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,zone,name,text")), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Display("/plugin/table.js", "style", "card")
			mdb.HashSelect(m, arg...).PushAction(cli.START, mdb.REMOVE)
			m.PushPodCmd(cmd, arg...)
		}},
	}})
}
