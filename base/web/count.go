package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const COUNT = "count"

func init() {
	Index.MergeCommands(ice.Commands{
		COUNT: &ice.Command{Name: "count hash auto", Help: "计数", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					count, limit, list := 0, 5, map[string]bool{}
					mdb.HashSelect(m.Spawn(kit.Dict(ice.MSG_FIELDS, mdb.Config(m, mdb.FIELD)))).Sort(mdb.TIME, "time_r").Table(func(value ice.Maps) {
						switch value[mdb.TYPE] {
						case ctx.ACTION:
							if list[value[mdb.NAME]] {
								break
							}
							if count++; count <= limit {
								list[value[mdb.NAME]] = true
								m.PushSearch(mdb.TYPE, ice.CMD, value)
							}
						}
					})
					count, limit = 0, 5
					mdb.HashSelect(m.Spawn(kit.Dict(ice.MSG_FIELDS, mdb.Config(m, mdb.FIELD)))).Sort(mdb.COUNT, "int_r").Table(func(value ice.Maps) {
						switch value[mdb.TYPE] {
						case ctx.ACTION:
							if list[value[mdb.NAME]] {
								break
							}
							if count++; count <= limit {
								list[value[mdb.NAME]] = true
								m.PushSearch(mdb.TYPE, ice.CMD, value)
							}
						}
					})
				}
			}},
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectUpdate(m, mdb.HashCreate(m), func(value ice.Map) { value[mdb.COUNT] = kit.Int(value[mdb.COUNT]) + 1 })
			}},
		}, ctx.CmdAction(), mdb.HashAction(mdb.SHORT, "type,name", mdb.FIELD, "time,hash,count,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Sort("type,name,text", "str", "str", "str")
		}},
	})
}
