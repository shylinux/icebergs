package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const EVENTS = "events"

func init() {
	const (
		SUBSCRIBE        = "subscribe"
		UNSUBSCRIBE      = "unsubscribe"
		SCAN             = "scan"
		SCANCODE_WAITMSG = "scancode_waitmsg"
		CLICK            = "click"
	)
	Index.MergeCommands(ice.Commands{
		EVENTS: {Help: "事件", Actions: ice.Actions{
			SUBSCRIBE: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.NAME, ice.Info.Titles)
				m.Option(mdb.TEXT, "欢迎光临")
				m.Option(mdb.ICONS, m.MergeLink(m.Resource(ice.Info.NodeIcon)))
				m.Cmdy(TEXT, web.LINK, m.MergeLink(nfs.PS))
			}},
			UNSUBSCRIBE: {Hand: func(m *ice.Message, arg ...string) {
			}},
			SCAN: {Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(SCAN, m.Option(ACCESS), arg[0])
				m.Options(ice.MSG_USERPOD, msg.Append(web.SPACE))
				link := m.Cmd(web.SHARE, mdb.CREATE, mdb.TYPE, web.FIELD, mdb.NAME, msg.Append(ctx.INDEX), mdb.TEXT, msg.Append(ctx.ARGS)).Option(web.LINK)
				m.Cmdy(TEXT, web.LINK, link, msg.Append(mdb.NAME), msg.Append(mdb.TEXT), msg.Append(mdb.ICONS))
			}},
			SCANCODE_WAITMSG: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TEXT, web.LINK, m.Option("ScanResult"))
			}},
			CLICK: {Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(MENU, m.Option(ACCESS), arg[0])
				m.Options(mdb.ICONS, msg.Append(mdb.ICONS), mdb.NAME, msg.Append(mdb.NAME), mdb.TEXT, kit.Select(msg.Append(ctx.INDEX), msg.Append(mdb.TEXT)))
				if m.Option(mdb.ICONS) == "" && msg.Append(ctx.INDEX) != "" {
					m.Search(msg.Append(ctx.INDEX), func(key string, cmd *ice.Command) {
						if cmd.Icon != "" {
							m.Option(mdb.ICONS, m.MergeLink(m.Resource(cmd.Icon)))
						}
					})
				}
				if msg.Append(ctx.INDEX) == "" {
					m.Cmdy(TEXT, web.LINK, m.MergeLink(nfs.PS))
				} else {
					m.Cmdy(TEXT, web.LINK, m.MergePodCmd("", msg.Append(ctx.INDEX), kit.Split(msg.Append(ctx.ARGS))))
				}
			}},
		}},
	})
}
