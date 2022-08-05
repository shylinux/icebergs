package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor zone id auto insert test page", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone=数据结构 type=go name=hi text=hello path file line", Help: "添加"},
			INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessCommand(m, INNER, m.OptionSplit("path,file,line"), arg...)
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			mdb.ZoneSelectPage(m, arg...).PushAction(kit.Select(mdb.REMOVE, INNER, len(arg) > 0))
		}},
	})
}
