package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _river_key(m *ice.Message, key ...ice.Any) string {
	return kit.Keys(mdb.HASH, m.Option(ice.MSG_RIVER), kit.Simple(key))
}
func _river_list(m *ice.Message) {
	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(mdb.TYPE) {
		case web.RIVER: // 共享群组
			m.Option(ice.MSG_TITLE, msg.Append(mdb.NAME))
			m.Option(ice.MSG_RIVER, msg.Append(RIVER))
			m.Option(ice.MSG_STORM, msg.Append(STORM))

			if m.Conf(RIVER, _river_key(m)) == "" {
				break // 虚拟群组
			}
			if msg.Cmd(OCEAN, m.Option(ice.MSG_USERNAME)).Append(aaa.USERNAME) == "" {
				msg.Cmd(OCEAN, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME)) // 加入群组
			}

		case web.STORM: // 共享应用
			m.Option(ice.MSG_TITLE, msg.Append(mdb.NAME))
			m.Option(ice.MSG_STORM, msg.Append(STORM))
			m.Option(ice.MSG_RIVER, "_share")
			return

		case web.FIELD: // 共享命令
			m.Option(ice.MSG_TITLE, msg.Append(mdb.NAME))
			m.Option(ice.MSG_RIVER, "_share")
			return
		}
	}

	m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, ice.OptionFields(mdb.HASH, mdb.NAME)).Tables(func(value ice.Maps) {
		m.Cmd(mdb.SELECT, m.PrefixKey(), kit.Keys(mdb.HASH, value[mdb.HASH], OCEAN), mdb.HASH, m.Option(ice.MSG_USERNAME)).Tables(func(val ice.Maps) {
			m.Push("", value, []string{mdb.HASH, mdb.NAME}, val)
		})
	})
}

const (
	RIVER_CREATE = "river.create"
)
const RIVER = "river"

func init() {
	Index.MergeCommands(ice.Commands{
		RIVER: {Name: "river hash auto create", Help: "群组", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.CommandKey())
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case cli.START, "创建空间":
					m.Cmdy(web.DREAM, mdb.INPUTS, arg)
					return
				}

				switch arg[0] {
				case nfs.TEMPLATE:
					m.Push(nfs.TEMPLATE, ice.BASE)
				case aaa.USERROLE:
					m.Push(aaa.USERROLE, aaa.VOID, aaa.TECH, aaa.ROOT)
				case aaa.USERNAME:
					m.Cmdy(aaa.USER).Cut(aaa.USERNAME, aaa.USERNICK, aaa.USERZONE)
				default:
					mdb.HashInputs(m, arg)
				}
			}},
			mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello template=base", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m, arg)
				gdb.Event(m, RIVER_CREATE, RIVER, m.Option(ice.MSG_RIVER, h), arg)
				m.Result(h)
			}},
			cli.START: {Name: "start name=hi repos template", Help: "创建空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.Space(m, m.Option(ice.POD)), web.DREAM, cli.START, arg)
			}},
			aaa.INVITE: {Name: "invite", Help: "添加设备", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("publish", ice.CONTEXTS)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,template"), web.ApiAction("/river")), Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
				m.RenderStatusUnauthorized()
				return // 没有登录
			}
			if len(arg) == 0 {
				m.Option(MENUS, m.Config(MENUS))
				_river_list(m)
				return // 群组列表
			}
			if len(arg) == 2 && arg[1] == STORM {
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy(arg[1], arg[2:])
				return // 应用列表
			}
			if !aaa.Right(m, RIVER, arg) {
				m.RenderStatusForbidden()
				return // 没有授权
			}

			if command := m.Commands(kit.Select("", arg, 1)); command != nil {
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy(arg[1], arg[2:])

			} else if action := m.Actions(kit.Select("", arg, 1)); action != nil {
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy("", arg[1:])

			} else {
				m.Cmdy(RIVER, arg)
			}
		}},
	})
}
