package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FLOWS = "flows"

func init() {
	Index.MergeCommands(ice.Commands{
		FLOWS: {Name: "flows refresh", Help: "工作流", Icon: "Automator.png", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.ShortKey(), "", mdb.HASH, m.OptionSimple(mdb.ZONE))
				m.Cmd(FLOWS, mdb.INSERT, mdb.NAME, m.Option(mdb.ZONE))
			}},
			mdb.INSERT: {Name: "insert name space index args", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.ShortKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.Config(m, mdb.FIELDS)))
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.ShortKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, m.ShortKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.HASH), arg)
			}},
		}, mdb.ExportHashAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,zone", mdb.FIELDS, "time,hash,name,space,index,args,prev,from,status")), Hand: func(m *ice.Message, arg ...string) {
			if arg = kit.Slice(arg, 0, 2); len(arg) == 0 || arg[0] == "" {
				mdb.HashSelect(m).Action(mdb.CREATE)
			} else {
				m.Fields(len(arg)-1, mdb.Config(m, mdb.FIELDS), mdb.DETAIL)
				m.Cmdy(mdb.SELECT, m.ShortKey(), kit.KeyHash(arg[0]), mdb.HASH, arg[1:])
				m.PushAction("addnext", "addto", mdb.RENAME, mdb.PLUGIN, mdb.DELETE)
			}
			m.Display("").DisplayCSS("")
		}},
	})
}
