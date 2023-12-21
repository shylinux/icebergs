package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const IFRAME = "iframe"

func init() {
	Index.MergeCommands(ice.Commands{
		IFRAME: {Name: "iframe hash@key auto", Help: "浏览器", Icon: "Safari.png", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.NAME:
					m.Push(arg[0], web.UserWeb(m).Host)
				case mdb.LINK:
					m.Push(arg[0], m.Option(ice.MSG_USERWEB))
					m.Copy(m.Cmd(web.SPIDE).CutTo(web.CLIENT_URL, arg[0]))
				case mdb.HASH:
					m.Cmd(mdb.SEARCH, mdb.FOREACH, "", "type,name,text", func(value ice.Maps) {
						kit.If(value[mdb.TYPE] == web.LINK, func() {
							m.Push(arg[0], value[mdb.TEXT]).Push(mdb.NAME, value[mdb.NAME])
						})
					})
				}
			}},
			mdb.CREATE: {Name: "create type name link", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessRewrite(mdb.HASH, mdb.HashCreate(m, mdb.TYPE, web.LINK, mdb.NAME, kit.ParseURL(m.Option(web.LINK)).Host, m.OptionSimple()))
			}},
			web.OPEN: {Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(m.Option(web.LINK)) }},
			ice.APP: {Help: "本机", Icon: "bi bi-browser-chrome", Hand: func(m *ice.Message, arg ...string) {
				defer m.ProcessHold()
				if h := kit.Select(m.Option(mdb.HASH), arg, 0); h != "" {
					cli.Opens(m, m.Cmd("", h).Append(mdb.LINK))
				} else {
					cli.Opens(m, "Safari.app")
				}
			}},
			FAVOR_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], web.LINK)
				default:
					if m.Option(mdb.TYPE) == web.LINK {
						switch arg[0] {
						case mdb.NAME:
							m.Push(arg[0], web.UserWeb(m).Host)
						case mdb.TEXT:
							m.Push(arg[0], m.Option(ice.MSG_USERWEB))
							m.Copy(m.Cmd(web.SPIDE).CutTo(web.CLIENT_URL, arg[0]))
						}
					}
				}
			}},
			FAVOR_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == web.LINK, func() { m.PushButton(kit.Dict(m.CommandKey(), "网页")) })
			}},
			FAVOR_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == web.LINK, func() { ctx.ProcessField(m, m.PrefixKey(), m.Option(mdb.TEXT)) })
			}},
		}, FavorAction(), mdb.HashAction(mdb.SHORT, web.LINK, mdb.FIELD, "time,hash,type,name,link")), Hand: func(m *ice.Message, arg ...string) {
			list := []string{m.MergePodCmd("", web.WIKI_PORTAL, log.DEBUG, m.Option(log.DEBUG))}
			list = append(list, m.MergePodCmd("", web.CHAT_PORTAL, log.DEBUG, m.Option(log.DEBUG)))
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				for _, link := range list {
					if u := kit.ParseURL(link); u != nil {
						m.Push("", kit.Dict(mdb.TIME, m.Time(), mdb.HASH, kit.Hashs(link), mdb.TYPE, web.LINK, mdb.NAME, u.Path, web.LINK, link))
					}
				}
				m.PushAction(web.OPEN, mdb.REMOVE)
				list := kit.List(mdb.CREATE)
				kit.If(m.Length() > 0, func() { list = append(list, mdb.PRUNES) })
				kit.If(web.IsLocalHost(m), func() { list = append(list, ice.APP) })
				m.Action(list...)
			} else {
				kit.If(m.Length() == 0, func() {
					for _, link := range list {
						if arg[0] == kit.Hashs(link) {
							m.Append(web.LINK, link)
							return
						}
					}
					m.Append(web.LINK, arg[0])
				})
				m.Action(web.FULL, web.OPEN, ice.APP).StatusTime(m.AppendSimple(web.LINK))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}
