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
		IFRAME: {Name: "iframe hash@key auto safari", Icon: "usr/icons/Safari.png", Help: "浏览器", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd("").Length() == 0 {
					m.Cmd(web.SPIDE, ice.OptionFields(web.CLIENT_NAME, web.CLIENT_ORIGIN), func(value ice.Maps) {
						if kit.IsIn(value[web.CLIENT_NAME], "ops", "dev", "com", "shy") {
							m.Cmd("", mdb.CREATE, kit.Dict(mdb.NAME, value[web.CLIENT_NAME], web.LINK, value[web.CLIENT_ORIGIN]))
						}
					})
				}
			}},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessRewrite(mdb.HASH, mdb.HashCreate(m, mdb.TYPE, web.LINK, mdb.NAME, kit.ParseURL(m.Option(web.LINK)).Host, m.OptionSimple()))
			}},
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
							m.Push(arg[0], value[mdb.TEXT])
							m.Push(mdb.NAME, value[mdb.NAME])
						})
					})
				}
			}},
			FAVOR_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], web.LINK)
				default:
					if m.Option(mdb.TYPE) != "" && m.Option(mdb.TYPE) != web.LINK {
						return
					}
					switch arg[0] {
					case mdb.NAME:
						m.Push(arg[0], web.UserWeb(m).Host)
					case mdb.LINK, ctx.ARGS:
						m.Push(arg[0], m.Option(ice.MSG_USERWEB))
						m.Copy(m.Cmd(web.SPIDE).CutTo(web.CLIENT_URL, arg[0]))
					}
				}
			}},
			FAVOR_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(arg[1] == web.LINK, func() { m.PushButton(IFRAME, mdb.REMOVE) })
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
			web.OPEN: {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessOpen(m, m.Option(web.LINK)) }},
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.CREATE, kit.Dict(web.LINK, m.MergePod(m.Option(mdb.NAME))))
			}},
			"safari": {Help: "本机", Hand: func(m *ice.Message, arg ...string) {
				if h := kit.Select(m.Option(mdb.HASH), arg, 0); h == "" {
					cli.Opens(m, "Safari.app")
				} else {
					cli.Opens(m, m.Cmd("", h).Append(mdb.LINK))
				}
			}},
		}, mdb.HashAction(mdb.SHORT, web.LINK, mdb.FIELD, "time,hash,type,name,link"), FavorAction()), Hand: func(m *ice.Message, arg ...string) {
			list := []string{m.MergePodCmd("", "web.wiki.portal", log.DEBUG, m.Option(log.DEBUG))}
			list = append(list, web.MergeLink(m, "/chat/portal/", ice.POD, m.Option(ice.MSG_USERPOD), log.DEBUG, m.Option(log.DEBUG)))
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				for _, link := range list {
					if u := kit.ParseURL(link); u != nil {
						m.Push("", kit.Dict(mdb.TIME, m.Time(), mdb.HASH, kit.Hashs(link), mdb.TYPE, web.LINK, mdb.NAME, u.Path, web.LINK, link))
					}
				}
				if m.Length() == 0 {
					m.Action(mdb.CREATE)
				} else {
					m.PushAction(web.OPEN, mdb.REMOVE).Action(mdb.CREATE, mdb.PRUNES)
				}
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
				m.Action(web.FULL, web.OPEN).StatusTime(m.AppendSimple(web.LINK))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}
