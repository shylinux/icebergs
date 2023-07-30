package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FLOWS = "flows"

func init() {
	Index.MergeCommands(ice.Commands{
		FLOWS: {Name: "flows zone hash auto", Help: "工作流", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert space index* args", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.PrefixKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.Config(m, mdb.FIELDS)))
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.PrefixKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, m.PrefixKey(), kit.KeyHash(m.Option(mdb.ZONE)), mdb.HASH, m.OptionSimple(mdb.HASH), arg)
			}},
			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.Option(ctx.INDEX), kit.Split(m.Option(ctx.ARGS)), arg...)
				if !kit.HasPrefixList(arg, ice.RUN) {
					m.Push("style", "float")
				}
			}},
		}, ctx.CmdAction(), mdb.HashAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,zone", mdb.FIELDS, "time,hash,space,index,args,prev,from,status")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				mdb.HashSelect(m).Action(mdb.CREATE)
			} else {
				arg = kit.Slice(arg, 0, 2)
				m.Fields(len(arg)-1, mdb.Config(m, mdb.FIELDS), mdb.DETAIL)
				m.Cmdy(mdb.SELECT, m.PrefixKey(), kit.KeyHash(arg[0]), mdb.HASH, arg[1:])
				m.Table(func(value ice.Maps) {
					switch value[mdb.STATUS] {
					case "done":
						m.PushButton(mdb.PLUGIN, mdb.DELETE)
					case "fail":
						m.PushButton(mdb.PLUGIN, mdb.DELETE)
					default:
						m.PushButton(mdb.PLUGIN, mdb.DELETE)
					}
				}).Action(mdb.INSERT).StatusTimeCount()
			}
			m.Display("")
		}},
	})
}
