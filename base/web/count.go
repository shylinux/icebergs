package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const COUNT = "count"

func init() {
	Index.MergeCommands(ice.Commands{
		COUNT: &ice.Command{Name: "count hash auto", Help: "计数", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectUpdate(m, mdb.HashCreate(m), func(value ice.Map) { value[mdb.COUNT] = kit.Int(value[mdb.COUNT]) + 1 })
			}},
		}, mdb.HashAction(mdb.SHORT, "type,name", mdb.FIELD, "time,hash,count,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Sort("type,name,text", "str", "str", "str")
		}},
	})
}
