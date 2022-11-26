package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
)

const KEYBOARD = "keyboard"

func init() {
	Index.MergeCommands(ice.Commands{
		KEYBOARD: {Name: "keyboard hash auto", Help: "键盘", Actions: ice.MergeActions(ice.Actions{
			web.SPACE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.SPACE, m.Option(web.SPACE), arg) }},
		}, mdb.HashAction(mdb.FIELD, "time,hash,space,index,input")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				msg := m.Cmd(ctx.COMMAND, m.Append(ctx.INDEX))
				m.Push(mdb.META, msg.Append(mdb.META))
				m.Push(mdb.LIST, msg.Append(mdb.LIST))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

func KeyboardAction() ice.Actions {
	return ice.Actions{
		KEYBOARD: {Hand: func(m *ice.Message, arg ...string) {
			hash := m.Cmdx("web.chat.keyboard", mdb.CREATE, web.SPACE, m.Option(ice.MSG_DAEMON), ctx.INDEX, m.Option(ctx.INDEX), "input", "")
			link := tcp.PublishLocalhost(m, web.MergePodCmd(m, "", "web.chat.keyboard", mdb.HASH, hash))
			m.Push(mdb.NAME, link).PushQRCode(mdb.TEXT, link)
		}},
	}
}
