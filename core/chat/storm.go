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
		STORM: {Name: "storm hash id auto", Help: "应用", Actions: ice.Actions{
			mdb.CREATE: {Name: "create name=hi text=hello", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m), mdb.HASH, arg)
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m), mdb.HASH, mdb.HASH, m.Option(ice.MSG_STORM))
			}},
			mdb.INSERT: {Name: "insert hash space index args style display", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _storm_key(m), mdb.LIST, arg)
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, RIVER, _storm_key(m), mdb.LIST, arg, "deleted", ice.TRUE)
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == mdb.ID {
					m.Cmdy(mdb.MODIFY, RIVER, _storm_key(m), mdb.LIST, arg)
				} else {
					m.Cmdy(mdb.MODIFY, RIVER, _river_key(m), mdb.HASH, mdb.HASH, m.Option(ice.MSG_STORM), arg)
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_STORM) == "" {
				m.Cmdy(mdb.SELECT, RIVER, _river_key(m), mdb.HASH, ice.OptionFields("time,hash,icon,name,text,count"))
				m.Action(mdb.CREATE)
			} else if len(arg) == 0 || kit.Int(arg[0]) > 0 {
				m.Cmdy(mdb.SELECT, RIVER, _storm_key(m), mdb.LIST, mdb.ID, arg, ice.OptionFields("time,id,space,icon,index,args,style,display,deleted")).SortInt(mdb.ID)
				m.Table(func(value ice.Maps) {
					msg := m.Cmd(ctx.COMMAND, value[mdb.INDEX])
					m.Push(mdb.NAME, msg.Append(mdb.NAME))
					m.Push(mdb.HELP, msg.Append(mdb.HELP))
					m.Push(mdb.META, msg.Append(mdb.META))
					m.Push(mdb.LIST, msg.Append(mdb.LIST))
					m.Push("_help", msg.Append("_help"))
				})
				m.Action(mdb.INSERT)
			} else if aaa.Right(m, arg) {
				m.Push(ctx.INDEX, arg[0])
			}
		}},
	})
}
