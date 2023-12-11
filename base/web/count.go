package web

import (
	"strings"

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
						if value["location"] == "" {
							location := kit.Format(kit.Value(SpideGet(m, "http://opendata.baidu.com/api.php?co=&resource_id=6006&oe=utf8", "query", value[mdb.NAME]), "data.0.location"))
							mdb.HashModify(m, mdb.HASH, value[mdb.HASH], "location", location)
							toast(location, index, m.Length())
							m.Sleep300ms()
						}
					})
					return nil
				})
			}},
		}, mdb.HashAction(mdb.LIMIT, 1000, mdb.LEAST, 500, mdb.SHORT, "type,name", mdb.FIELD, "time,hash,count,location,type,name,text", mdb.SORT, "type,name,text,location")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
			stat := map[string]int{}
			m.Table(func(value ice.Maps) {
				for _, agent := range []string{"GoModuleMirror"} {
					if strings.Contains(value[mdb.TEXT], agent) {
						stat[agent]++
					}
				}
			})
			m.StatusTimeCount(stat)
		}},
	})
}
