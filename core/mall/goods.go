package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)
const GOODS = "goods"

func init() {
	Index.MergeCommands(ice.Commands{
		GOODS: {Name: "goods hash auto", Help: "商品", Actions: ice.MergeActions(ice.Actions{
			
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || arg[0] == "" {
				m.Action(mdb.CREATE)
			}
		}},
	})	
}