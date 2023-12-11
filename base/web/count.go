package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const COUNT = "count"

func init() {
	Index.MergeCommands(ice.Commands{
		COUNT: &ice.Command{Name: "count hash auto location", Help: "计数", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectUpdate(m, mdb.HashCreate(m), func(value ice.Map) { value[mdb.COUNT] = kit.Int(value[mdb.COUNT]) + 1 })
			}},
			"location": {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelects(m).Sort(mdb.COUNT, ice.INT_R)
				GoToast(m, "", func(toast func(string, int, int)) []string {
					m.Table(func(index int, value ice.Maps) {
						location := kit.Format(kit.Value(SpideGet(m, "http://opendata.baidu.com/api.php?query=%s&co=&resource_id=6006&oe=utf8", value["name"]), "data.0.location"))
						toast(location, index, m.Length())
						mdb.HashModify(m, mdb.HASH, value[mdb.HASH], "location", location)
						m.Sleep("500ms")
					})
					return nil
				})
			}},
		}, mdb.HashAction(mdb.LIMIT, 1000, mdb.LEAST, 500, mdb.SHORT, "type,name", mdb.FIELD, "time,hash,count,type,name,text", mdb.SORT, "type,name,text,location"))},
	})
}
