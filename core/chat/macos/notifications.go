package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

const NOTIFICATIONS = "notifications"

func init() {
	Index.MergeCommands(ice.Commands{
		NOTIFICATIONS: {Actions: ice.MergeActions(ice.Actions{
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				Notify(m, "Launchpad.png", m.Option(mdb.NAME), "空间创建成功", ctx.INDEX, web.CHAT_IFRAME, ctx.ARGS, m.MergePod(m.Option(mdb.NAME)))
			}},
			"read": {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.MODIFY, m.PrefixKey(), "", mdb.LIST, m.OptionSimple(mdb.ID), mdb.STATUS, m.ActionKey())
			}},
		}, gdb.EventsAction(web.DREAM_CREATE), mdb.ListAction(mdb.FIELD, "time,id,status,icon,name,text,space,index,args")), Hand: func(m *ice.Message, arg ...string) {
			mdb.ListSelect(m, arg...).Display("")
		}},
	})
}
func Notify(m *ice.Message, icon, name, text string, arg ...string) {
	nfs.Exists(m, ice.USR_ICONS+icon, func(p string) { icon = p })
	m.Cmd(NOTIFICATIONS, mdb.INSERT, mdb.ICON, icon, mdb.NAME, name, mdb.TEXT, text, arg)
}
