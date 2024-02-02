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
		FAVOR: {Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create zone*=数据结构"},
			mdb.INSERT: {Name: "insert zone*=数据结构 type=go name*=hi text*=hello path file line"},
			XTERM: {Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, "", func() []string {
					return mdb.ZoneSelects(m.Spawn(), m.Option(mdb.ZONE), m.Option(mdb.ID)).OptionSplit(mdb.TYPE, mdb.NAME, mdb.TEXT)
				}, arg...)
			}},
			INNER: {Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, "", func() []string {
					return mdb.ZoneSelects(m.Spawn(), m.Option(mdb.ZONE), m.Option(mdb.ID)).OptionSplit(nfs.PATH, nfs.FILE, nfs.LINE)
				}, arg...)
			}},
		}, mdb.PageZoneAction(mdb.FIELDS, "time,id,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.PageZoneSelect(m, arg...); len(arg) > 0 && arg[0] != "" {
				m.Table(func(value ice.Maps) { m.PushButton(kit.Select(INNER, XTERM, value[nfs.FILE] == "")) })
				m.Option(ctx.STYLE, arg[0])
			}
		}},
	})
}
