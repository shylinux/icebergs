package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _storm_key(m *ice.Message, key ...ice.Any) string {
	return _river_key(m, mdb.HASH, m.Option(ice.MSG_STORM), kit.Keys(key))
}

const STORM = "storm"

func init() {
	Index.MergeCommands(ice.Commands{
		STORM: {Name: "storm hash id auto insert create", Help: "应用", Actions: ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {}},
			mdb.CREATE: {Name: "create name=hi text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m), mdb.HASH, arg)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m), mdb.HASH, mdb.HASH, m.Option(ice.MSG_STORM))
			}},
			mdb.INSERT: {Name: "insert hash space index args style display", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _storm_key(m), mdb.LIST, arg)
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == mdb.ID {
					m.Cmdy(mdb.MODIFY, RIVER, _storm_key(m), mdb.LIST, arg)
				} else {
					m.Cmdy(mdb.MODIFY, RIVER, _river_key(m), mdb.HASH, mdb.HASH, m.Option(ice.MSG_STORM), arg)
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_STORM) == "" { // 应用列表
				m.OptionFields("time,hash,name,text,count")
				m.Cmdy(mdb.SELECT, RIVER, _river_key(m), mdb.HASH)

			} else if len(arg) == 0 || kit.Int(arg[0]) > 0 { // 工具列表
				m.OptionFields("time,id,space,index,args,style,display")
				m.Cmdy(mdb.SELECT, RIVER, _storm_key(m), mdb.LIST, mdb.ID, arg)

			} else if aaa.Right(m, arg[0]) { // 静态群组
				m.Push(ctx.INDEX, arg[0])
			}
		}},
	})
}
