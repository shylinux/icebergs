package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const IFRAME = "iframe"

func init() {
	Index.MergeCommands(ice.Commands{
		IFRAME: {Name: "iframe hash auto", Help: "浏览器", Actions: ice.MergeActions(ice.Actions{
			FAVOR_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], web.LINK)
				default:
					if m.Option(mdb.TYPE) != web.LINK {
						return
					}
					switch arg[0] {
					case mdb.NAME:
						m.Push(arg[0], web.OptionUserWeb(m).Host)
					case mdb.TEXT:
						m.Push(arg[0], m.Option(ice.MSG_USERWEB))
					}
				}
			}},
			FAVOR_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[1] {
				case web.LINK:
					m.PushButton(IFRAME, web.OPEN, mdb.REMOVE)
				}
			}},
			FAVOR_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TYPE) != web.LINK {
					return
				}
				switch kit.Select("", arg, 1) {
				case web.OPEN:
					ctx.ProcessOpen(m, m.Option(mdb.TEXT))
				default:
					ctx.ProcessField(m, m.PrefixKey(), []string{m.Option(mdb.TEXT)}, arg...)
				}
			}},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, mdb.TYPE, web.LINK, mdb.NAME, kit.ParseURL(m.Option(web.LINK)).Host, m.OptionSimple())
			}},
			web.OPEN: {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessOpen(m, m.Option(web.LINK)) }},
		}, mdb.HashAction(mdb.SHORT, web.LINK, mdb.FIELD, "time,hash,type,name,link"), FavorAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.PushAction(web.OPEN, mdb.REMOVE).Action(mdb.CREATE, mdb.PRUNES)
			} else {
				if m.Length() == 0 {
					m.Append(web.LINK, arg[0])
				}
				m.Action(web.FULL, web.OPEN).StatusTime(m.AppendSimple(web.LINK))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}
