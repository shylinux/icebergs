package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _storm_key(m *ice.Message, key ...ice.Any) string {
	return _river_key(m, STORM, mdb.HASH, kit.Keys(key))
}

const STORM = "storm"

func init() {
	Index.MergeCommands(ice.Commands{
		STORM: {Name: "storm hash id auto insert create", Help: "应用", Actions: ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if ctx.Inputs(m, arg[0]) {
					return
				}
				switch arg[0] {
				case mdb.HASH:
					m.Cmdy("", ice.OptionFields("hash,name"))
				}
			}},
			mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, STORM), mdb.HASH, arg)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m, STORM), mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.INSERT: {Name: "insert hash space index", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _storm_key(m, m.Option(mdb.HASH)), mdb.LIST, arg[2:])
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.ID) == "" {
					m.Cmdy(mdb.MODIFY, RIVER, _river_key(m, STORM), mdb.HASH, m.OptionSimple(mdb.HASH), arg)
				} else {
					m.Cmdy(mdb.MODIFY, RIVER, _storm_key(m, m.Option(mdb.HASH)), mdb.LIST, m.OptionSimple(mdb.ID), arg)
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 { // 应用列表
				m.OptionFields("time,hash,type,name,count")
				m.Cmdy(mdb.SELECT, RIVER, _river_key(m, STORM), mdb.HASH).Sort(mdb.NAME)
				m.PushAction(mdb.REMOVE)
				return
			}

			m.OptionFields("time,id,space,index,args,style,display")
			msg := m.Cmd(mdb.SELECT, RIVER, _storm_key(m, arg[0]), mdb.LIST, mdb.ID, kit.Select("", arg, 1))
			if msg.Length() == 0 && len(arg) > 1 { // 虚拟群组
				msg.Push(ctx.INDEX, arg[1])
			}

			if len(arg) > 2 && arg[2] == ice.RUN { // 执行命令
				m.Cmdy(web.Space(m, kit.Select(m.Option(ice.POD), msg.Append(web.SPACE))), msg.Append(ctx.INDEX), arg[3:])
				return
			}

			if m.Copy(msg); len(arg) > 1 { // 命令插件
				m.Tables(func(value ice.Maps) { m.Cmdy(web.Space(m, value[web.SPACE]), ctx.COMMAND, value[ctx.INDEX]) })
				m.ProcessField(arg[0], arg[1], ice.RUN)
			}
		}},
	})
}
