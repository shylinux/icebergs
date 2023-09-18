package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FLOWS = "flows"

func init() {
	Index.MergeCommands(ice.Commands{
		FLOWS: {Name: "flows zone hash auto", Icon: "usr/icons/Automator.png", Help: "工作流", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					mdb.HashSelect(m.Spawn(ice.OptionFields(""))).Table(func(value ice.Maps) {
						m.PushSearch(mdb.TYPE, "", mdb.NAME, value[mdb.ZONE], mdb.TEXT, value[mdb.ZONE], value)
					})
				}
			}},
			mdb.INSERT: {Name: "insert name space index* args", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.PrefixKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.Config(m, mdb.FIELDS)))
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.PrefixKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, m.PrefixKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.HASH), arg)
			}},
		}, ctx.CmdAction(), mdb.ExportHashAction(), mdb.HashAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,zone", mdb.FIELDS, "time,hash,name,space,index,args,prev,from,status")), Hand: func(m *ice.Message, arg ...string) {
			if arg = kit.Slice(arg, 0, 2); len(arg) == 0 || arg[0] == "" {
				mdb.HashSelect(m)
			} else {
				m.Fields(len(arg)-1, mdb.Config(m, mdb.FIELDS), mdb.DETAIL)
				m.Cmdy(mdb.SELECT, m.PrefixKey(), kit.KeyHash(arg[0]), mdb.HASH, arg[1:])
				m.PushAction(mdb.PLUGIN, mdb.DELETE).StatusTimeCount()
			}
			m.Display("")
		}},
	})
}
