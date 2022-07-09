package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _storm_key(m *ice.Message, key ...ice.Any) string {
	return _river_key(m, STORM, mdb.HASH, kit.Keys(key))
}

const STORM = "storm"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		STORM: {Name: "storm hash id auto insert create", Help: "应用", Actions: ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if ctx.Inputs(m, arg[0]) {
					return
				}
				switch arg[0] {
				case mdb.HASH:
					m.Cmdy(STORM, ice.OptionFields("hash,name"))
				}
			}},
			mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, STORM), mdb.HASH, arg)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m, STORM), mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.INSERT: {Name: "insert hash pod ctx cmd help", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _storm_key(m, m.Option(mdb.HASH)), mdb.LIST, arg[2:])
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.ID) == "" {
					m.Cmdy(mdb.MODIFY, RIVER, _river_key(m, STORM), mdb.HASH, m.OptionSimple(mdb.HASH), arg)
				} else {
					m.Cmdy(mdb.MODIFY, RIVER, _storm_key(m, m.Option(mdb.HASH)), mdb.LIST, m.OptionSimple(mdb.ID), arg)
				}
			}},
			mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.ID) == "" {
					return
				}
				msg := m.Cmd(STORM, m.Option(mdb.HASH), m.Option(mdb.ID))
				cmd := kit.Keys(msg.Append(ice.CTX), msg.Append(ice.CMD))
				_action_domain(m, cmd, m.Option(mdb.HASH))
				m.Cmdy(cmd, mdb.EXPORT)
			}},
			mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.ID) == "" {
					return
				}
				msg := m.Cmd(STORM, m.Option(mdb.HASH), m.Option(mdb.ID))
				cmd := kit.Keys(msg.Append(ice.CTX), msg.Append(ice.CMD))
				_action_domain(m, cmd, m.Option(mdb.HASH))
				m.Cmdy(cmd, mdb.IMPORT)
			}},
			SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) { _header_share(m, arg...) }},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 { // 应用列表
				m.OptionFields("time,hash,type,name,count")
				m.Cmdy(mdb.SELECT, RIVER, _river_key(m, STORM), mdb.HASH).Sort(mdb.NAME)
				m.PushAction(mdb.REMOVE)
				return
			}

			m.OptionFields("time,id,pod,ctx,cmd,arg,display,style")
			msg := m.Cmd(mdb.SELECT, RIVER, _storm_key(m, arg[0]), mdb.LIST, mdb.ID, kit.Select("", arg, 1))
			if msg.Length() == 0 && len(arg) > 1 { // 虚拟群组
				msg.Push(ice.CMD, arg[1])
			}

			if len(arg) > 2 && arg[2] == ice.RUN { // 执行命令
				m.Cmdy(m.Space(kit.Select(m.Option(ice.POD), msg.Append(ice.POD))), kit.Keys(msg.Append(ice.CTX), msg.Append(ice.CMD)), arg[3:])
				return
			}

			if m.Copy(msg); len(arg) > 1 { // 命令插件
				m.ProcessField(arg[0], arg[1], ice.RUN)
				m.Table(func(index int, value ice.Maps, head []string) {
					m.Cmdy(m.Space(value[ice.POD]), ctx.CONTEXT, value[ice.CTX], ctx.COMMAND, value[ice.CMD])
				})
			} else {
				m.PushAction(mdb.EXPORT, mdb.IMPORT)
			}
		}},
	}})
}
