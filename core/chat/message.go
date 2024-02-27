package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const MESSAGE = "message"

func init() {
	Index.MergeCommands(ice.Commands{
		MESSAGE: {Name: "message", Help: "聊天", Icon: "Messages.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				messageCreate(m, web.DREAM, "usr/icons/Launchpad.png")
				messageCreate(m, cli.SYSTEM, "usr/icons/System Settings.png")
				messageInsert(m, cli.SYSTEM, mdb.TYPE, "plug", ctx.INDEX, cli.RUNTIME)
			}},
			mdb.CREATE: {Name: "create type*=tech,void zone* icons* target"},
			mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, append(arg, aaa.AVATAR, aaa.UserInfo(m, "", aaa.AVATAR, aaa.AVATAR), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERNAME, m.Option(ice.MSG_USERNAME)))
				kit.If(mdb.HashSelectField(m, arg[0], "target"), func(p string) {
					m.Cmd(web.SPACE, p, MESSAGE, tcp.RECV, arg[1:])
				})
				mdb.HashSelectUpdate(m, arg[0], func(value ice.Map) {
					kit.Value(value, mdb.TIME, m.Time())
				})
			}},
			tcp.RECV: {Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, kit.Simple(mdb.ZONE, m.Option(ice.FROM_SPACE), web.SPACE, m.Option(ice.FROM_SPACE), arg, aaa.AVATAR, aaa.UserInfo(m, "", aaa.AVATAR, aaa.AVATAR), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERNAME, m.Option(ice.MSG_USERNAME)))
				mdb.HashSelectUpdate(m, m.Option(ice.FROM_SPACE), func(value ice.Map) {
					kit.Value(value, "target", m.Option(ice.FROM_SPACE))
				})
				mdb.HashSelectUpdate(m, m.Option(ice.FROM_SPACE), func(value ice.Map) {
					kit.Value(value, mdb.TIME, m.Time())
				})
			}},
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				messageInsert(m, web.DREAM, mdb.TYPE, "plug", ctx.INDEX, IFRAME, ctx.ARGS, web.S(m.Option(mdb.NAME)))
			}},
		}, web.DreamAction(), mdb.ZoneAction(
			mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,hash,type,zone,icons,target", mdb.FIELDS, "time,id,avatar,usernick,username,type,name,text,space,index,args",
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.ZoneSelect(m.Display("").Spawn(), arg...).Table(func(value ice.Maps) {
					if kit.IsIn(m.Option(ice.MSG_USERROLE), value[mdb.TYPE], aaa.TECH, aaa.ROOT) {
						m.PushRecord(value, mdb.Config(m, mdb.FIELD))
					}
				})
				m.Sort(mdb.TIME, "str_r")
			} else {
				mdb.ZoneSelect(m, arg...).Sort(mdb.ID, ice.INT)
			}
		}},
	})
}
func messageCreate(m *ice.Message, zone, icons string) {
	kit.Value(m.Target().Configs[m.CommandKey()].Value, kit.Keys(mdb.HASH, zone, mdb.META), kit.Dict(
		mdb.TIME, m.Time(), mdb.TYPE, aaa.TECH, mdb.ZONE, zone, mdb.ICONS, icons,
	))
}
func messageInsert(m *ice.Message, zone string, arg ...string) {
	m.Cmd("", mdb.INSERT, zone, arg, ice.Maps{ice.MSG_USERNICK: zone, ice.MSG_USERNAME: zone})
}
