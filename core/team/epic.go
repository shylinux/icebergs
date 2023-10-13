package team

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const EPIC = "epic"

func init() {
	Index.MergeCommands(ice.Commands{
		EPIC: {Help: "史记", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create time@date zone name"}, mdb.MODIFY: {Name: "modify time zone name"},
		}, mdb.ExportHashAction(mdb.FIELD, "time,hash,zone,name")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
				if span := kit.Time(m.Time()) - kit.Time(value[mdb.TIME]); span > 0 {
					m.Push(mdb.TEXT, nfs.Template(m, "gone.html", m.Options("days", int(time.Duration(span)/time.Hour/24), "from", kit.Split(value[mdb.TIME])[0])))
				} else {
					m.Push(mdb.TEXT, nfs.Template(m, "will.html", m.Options("days", -int(time.Duration(span)/time.Hour/24)+1, "from", kit.Split(value[mdb.TIME])[0])))
				}
			}).PushAction(mdb.MODIFY, mdb.REMOVE)
			web.PushPodCmd(m, "", arg...)
			ctx.DisplayTableCard(m)
		}},
	})
}
