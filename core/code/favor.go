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
			mdb.INSERT: {Name: "insert zone=数据结构 type=go name=hi text=hello path file line", Help: "添加"},
			INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				ctx.Process(m, m.ActionKey(), m.OptionSplit(nfs.PATH, nfs.FILE, nfs.LINE), arg...)
			}},
			XTERM: {Name: "xterm", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				ctx.Process(m, m.ActionKey(), m.OptionSimple(mdb.TYPE, mdb.NAME, mdb.TEXT), arg...)
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelectPage(m, arg...); len(arg) > 0 && arg[0] != "" {
				m.Tables(func(value ice.Maps) {
					m.PushButton(kit.Select(INNER, XTERM, value[mdb.TEXT] == "" || value[nfs.FILE] == ""))
				})
			}
		}},
	})
}
