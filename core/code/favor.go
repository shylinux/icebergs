package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor zone id auto insert page", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone*=数据结构 type=go name*=hi text*=hello path file line"},
			XTERM: {Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.ZoneSelects(m.Spawn(), m.Option(mdb.ZONE), m.Option(mdb.ID))
				ctx.Process(m, "", msg.OptionSplit(mdb.TYPE, mdb.NAME, mdb.TEXT), arg...)
			}},
			INNER: {Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.ZoneSelects(m.Spawn(), m.Option(mdb.ZONE), m.Option(mdb.ID))
				ctx.Process(m, "", msg.OptionSplit(nfs.PATH, nfs.FILE, nfs.LINE), arg...)
			}},
		}, mdb.PageZoneAction(mdb.FIELD, "time,id,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.PageZoneSelect(m, arg...); len(arg) > 0 && arg[0] != "" {
				m.Table(func(value ice.Maps) { m.PushButton(kit.Select(INNER, XTERM, value[nfs.FILE] == "")) }).Option(ctx.STYLE, arg[0])
			}
		}},
	})
}
