package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const MESSAGE = "message"

func init() {
	Index.MergeCommands(ice.Commands{
		MESSAGE: {Name: "message", Help: "消息", Icon: "Messages.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				messageCreate(m, web.DREAM, "usr/icons/Launchpad.png")
				messageCreate(m, cli.SYSTEM, "usr/icons/System Settings.png")
				messageInsert(m, cli.SYSTEM, mdb.TYPE, "plug", ctx.INDEX, cli.RUNTIME)
			}},
			mdb.CREATE: {Name: "create type*=tech,void name* icons*"},
			mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, append(arg, aaa.AVATAR, aaa.UserInfo(m, "", aaa.AVATAR, aaa.AVATAR),
					aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERNAME, m.Option(ice.MSG_USERNAME),
				))
			}},
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				messageInsert(m, web.DREAM, mdb.TYPE, "plug", ctx.INDEX, IFRAME, ctx.ARGS, web.S(m.Option(mdb.NAME)))
			}},
		}, web.DreamAction(), mdb.ZoneAction(
			mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,type,name,icons", mdb.FIELDS, "time,avatar,usernick,username,type,name,text,space,index,args",
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.ZoneSelect(m.Spawn(), arg...).Table(func(value ice.Maps) {
					if kit.IsIn(m.Option(ice.MSG_USERROLE), value[mdb.TYPE], aaa.TECH, aaa.ROOT) {
						m.PushRecord(value, mdb.Config(m, mdb.FIELD))
					}
				})
			} else {
				mdb.ZoneSelect(m, arg...)
			}
			m.Display("")
		}},
	})
}
func messageCreate(m *ice.Message, name, icon string) {
	kit.Value(m.Target().Configs[m.CommandKey()].Value, kit.Keys(mdb.HASH, name, "meta.time"), m.Time())
	kit.Value(m.Target().Configs[m.CommandKey()].Value, kit.Keys(mdb.HASH, name, "meta.type"), aaa.TECH)
	kit.Value(m.Target().Configs[m.CommandKey()].Value, kit.Keys(mdb.HASH, name, "meta.name"), name)
	kit.Value(m.Target().Configs[m.CommandKey()].Value, kit.Keys(mdb.HASH, name, "meta.icons"), icon)
}
func messageInsert(m *ice.Message, zone string, arg ...string) {
	m.Cmd("", mdb.INSERT, zone, arg, ice.Maps{ice.MSG_USERNAME: zone})
}
