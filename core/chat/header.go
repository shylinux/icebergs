package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _header_users(m *ice.Message, arg ...string) {
	if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
		return
	} else if m.Warn(m.Option(web.SHARE) != "", ice.ErrNotRight, "没有权限") {
		return
	} else {
		m.Cmdy(aaa.USER, mdb.MODIFY, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.ActionKey(), m.Option(m.ActionKey(), arg[0]))
	}
}
func _header_share(m *ice.Message, arg ...string) {
	if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, "没有登录") {
		return
	} else if kit.For(arg, func(k, v string) { m.Option(k, v) }); m.Option(mdb.LINK) == "" {
		m.Cmdy(web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN, arg)
	} else {
		m.Option(mdb.LINK, tcp.PublishLocalhost(m, m.Option(mdb.LINK)))
	}
	m.Push(mdb.NAME, m.Option(mdb.LINK)).PushQRCode(mdb.TEXT, m.Option(mdb.LINK))
}
func _header_check(m *ice.Message, arg ...string) bool {
	m.Option(ice.MAIN, mdb.Conf(m, web.SERVE, kit.Keym(ice.MAIN)))
	if m.Option(ice.CMD) == aaa.OFFER && m.Option(mdb.HASH) != "" {
		m.Cmd(aaa.OFFER, m.Option(mdb.HASH), func(value ice.Maps) {
			aaa.SessAuth(m, kit.Dict(aaa.USERNAME, value[aaa.EMAIL], aaa.USERROLE, aaa.VOID))
		})
	} else if m.Option(web.SHARE) != "" {
		m.Cmd(web.SHARE, m.Option(web.SHARE), ice.OptionFields(""), func(value ice.Maps) {
			if web.IsNotValidShare(m, value[mdb.TIME]) {
				return
			}
			switch value[mdb.TYPE] {
			case web.FIELD, web.STORM:
				aaa.SessAuth(m, kit.Dict(value))
			}
		})
	}
	return m.Option(ice.MSG_USERNAME) != ""
}

const (
	TITLE = "title"
	THEME = "theme"
	MENUS = "menus"

	HEADER_AGENT = "header.agent"
)
const HEADER = "header"

func init() {
	Index.MergeCommands(ice.Commands{
		HEADER: {Help: "标题栏", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				aaa.White(m, HEADER)
				aaa.Black(m, kit.Keys(HEADER, ctx.ACTION, mdb.CREATE))
				aaa.Black(m, kit.Keys(HEADER, ctx.ACTION, mdb.REMOVE))
				aaa.Black(m, kit.Keys(HEADER, ctx.ACTION, mdb.MODIFY))
				kit.If(mdb.HashSelect(m.Spawn()).Length() == 0, func() {
					m.Cmd("", mdb.CREATE, mdb.TYPE, cli.QRCODE, mdb.NAME, "扫码登录", mdb.ORDER, "1")
				})
			}},
			web.SHARE:      {Hand: _header_share},
			aaa.LOGIN:      {Hand: func(m *ice.Message, arg ...string) {}},
			aaa.LOGOUT:     {Hand: aaa.SessLogout},
			aaa.PASSWORD:   {Hand: _header_users},
			aaa.USERNICK:   {Hand: _header_users},
			aaa.LANGUAGE:   {Hand: _header_users},
			aaa.BACKGROUND: {Hand: _header_users},
			aaa.AVATAR:     {Hand: _header_users},
			aaa.THEME: {Hand: func(m *ice.Message, arg ...string) {
				if tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
					m.Cmd(cli.SYSTEM, "osascript", "-e", `tell app "System Events" to tell appearance preferences to set dark mode to `+
						kit.Select(ice.TRUE, ice.FALSE, kit.IsIn(kit.Select(html.LIGHT, arg, 0), html.LIGHT, html.WHITE)))
				}
			}},
			aaa.EMAIL: {Name: "email to subject content", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.TO) != aaa.UserEmail(m, "") && !aaa.Right(m, aaa.EMAIL, m.Option(nfs.TO)) {
					return
				}
				m.Options("volcano", web.UserHost(m), nfs.VERSION, web.RenderVersion(m))
				m.Options(ice.MSG_USERWEB, tcp.PublishLocalhost(m, m.Option(ice.MSG_USERWEB)))
				m.Options(ice.MSG_USERWEB, kit.MergeURL(m.Option(ice.MSG_USERWEB), web.SHARE, m.Cmdx(web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN)))
				m.Cmdy(aaa.EMAIL, aaa.SEND, arg, aaa.CONTENT, nfs.Template(m, "email.html"))
			}},
			ice.DEMO: {Help: "体验", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_USERROLE) == aaa.TECH {
					if mdb.Config(m, ice.DEMO) == ice.TRUE {
						mdb.Config(m, ice.DEMO, ice.FALSE)
					} else {
						mdb.Config(m, ice.DEMO, ice.TRUE)
						m.Cmd("", mdb.CREATE, mdb.TYPE, mdb.PLUGIN, mdb.NAME, "免登录体验", mdb.ORDER, "2", ctx.INDEX, HEADER, ctx.ARGS, ice.DEMO)
					}
					return
				}
				if mdb.Config(m, ice.DEMO) == ice.TRUE {
					web.RenderCookie(m, aaa.SessCreate(m, ice.Info.Username))
					m.Echo("login success")
				} else {
					m.Echo("hello world")
				}
			}},
			mdb.CREATE: {Name: "create type*=oauth,plugin,qrcode name* icons link order space index args", Hand: func(m *ice.Message, arg ...string) { mdb.HashCreate(m, m.OptionSimple()) }},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { mdb.HashRemove(m, m.OptionSimple(mdb.NAME)) }},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, m.OptionSimple(mdb.NAME), arg) }},
		}, web.ApiAction(), mdb.ImportantHashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,icons,type,link,order,space,index,args")), Hand: func(m *ice.Message, arg ...string) {
			m.Option("language.list", m.Cmd(nfs.DIR, nfs.TemplatePath(m, aaa.LANGUAGE)+nfs.PS, nfs.FILE).Appendv(nfs.FILE))
			m.Option("theme.list", m.Cmd(nfs.DIR, nfs.TemplatePath(m, aaa.THEME)+nfs.PS, nfs.FILE).Appendv(nfs.FILE))
			m.Option(nfs.REPOS, m.Cmdv(web.SPIDE, nfs.REPOS, web.CLIENT_URL))
			m.Option("icon.lib", mdb.Conf(m, ICON, kit.Keym(nfs.PATH)))
			m.Option(MENUS, mdb.Config(m, MENUS))
			m.Echo(mdb.Config(m, TITLE))
			mdb.HashSelect(m, arg...).Sort("order", "int")
			if m.Option(ice.MSG_USERROLE) == aaa.TECH {
				m.Action(mdb.CREATE, ice.DEMO)
			}
			kit.If(GetSSO(m), func(p string) {
				m.Push(mdb.TIME, m.Time()).Push(mdb.NAME, web.SERVE).Push(mdb.ICONS, nfs.USR_ICONS_ICEBERGS).Push(mdb.TYPE, "oauth").Push(web.LINK, p)
			})
			m.StatusTimeCount(kit.Dict(ice.DEMO, mdb.Config(m, ice.DEMO)))
			if gdb.Event(m, HEADER_AGENT); !_header_check(m, arg...) {
				return
			}
			msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
			kit.For([]string{aaa.USERNICK, aaa.LANGUAGE, aaa.EMAIL}, func(k string) { m.Option(k, msg.Append(k)) })
			kit.For([]string{aaa.AVATAR, aaa.BACKGROUND}, func(k string) { m.Option(k, web.RequireFile(m, msg.Append(k))) })
			kit.If(m.Option(aaa.LANGUAGE) == "", func() {
				m.Option(aaa.LANGUAGE, kit.Select("", "zh-cn", strings.Contains(m.Option(ice.MSG_USERUA), "zh_CN")))
				kit.If(m.Option(aaa.LANGUAGE) == "", func() { m.Option(aaa.LANGUAGE, kit.Select("", kit.Split(m.R.Header.Get(web.AcceptLanguage), ",;"), 0)) })
				kit.If(m.Option(aaa.LANGUAGE) == "", func() { m.Option(aaa.LANGUAGE, ice.Info.Lang) })
			})
		}},
	})
}
