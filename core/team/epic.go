package team

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const EPIC = "epic"

func init() {
	Index.MergeCommands(ice.Commands{
		EPIC: {Name: "epic hash list create", Help: "史记", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create time@date zone name"}, mdb.MODIFY: {Name: "modify time zone name"},
		}, mdb.ImportantHashAction(mdb.FIELD, "time,hash,zone,name")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
				if span := kit.Time(m.Time()) - kit.Time(value[mdb.TIME]); span > 0 {
					m.Push(mdb.TEXT, kit.Format(`已经 <span style="font-size:24px;color:red">%v</span> 天<br>距 %s<br>`, int(time.Duration(span)/time.Hour/24), kit.Split(value[mdb.TIME])[0]))
				} else {
					m.Push(mdb.TEXT, kit.Format(`还有 <span style="font-size:24px;color:green">%v</span> 天<br>距 %s<br>`, -int(time.Duration(span)/time.Hour/24)+1, kit.Split(value[mdb.TIME])[0]))
				}
			}).PushAction(mdb.MODIFY, mdb.REMOVE)
			web.PushPodCmd(m, "", arg...)
			ctx.DisplayTableCard(m)
		}},
	})
}
