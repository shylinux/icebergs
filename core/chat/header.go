package chat

import (
	"net/http"
	"path"
	"strings"
	"time"

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
	if m.WarnNotLogin(m.Option(ice.MSG_USERNAME) == "") {
		return
	} else if m.WarnNotRight(m.Option(web.SHARE) != "", "没有权限") {
		return
	} else {
		m.Cmdy(aaa.USER, mdb.MODIFY, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.ActionKey(), m.Option(m.ActionKey(), arg[0]))
	}
}
func _header_share(m *ice.Message, arg ...string) {
	if m.WarnNotLogin(m.Option(ice.MSG_USERNAME) == "", "没有登录") {
		return
	} else if m.WarnNotRight(m.Option(web.SHARE) != "", "没有权限") {
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
			}},
			web.SHARE:      {Hand: _header_share},
			aaa.LANGUAGE:   {Hand: _header_users},
			aaa.USERNICK:   {Hand: _header_users},
			aaa.AVATAR:     {Hand: _header_users},
			aaa.BACKGROUND: {Hand: _header_users},
			web.UPLOAD: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				if m.WarnNotLogin(m.Option(ice.MSG_USERNAME) == "") {
					return
				}
				up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
				m.Cmdy(web.CACHE, web.WATCH, m.Option(mdb.HASH), path.Join("usr/avatar/", m.Option(ice.MSG_USERNAME)+"."+kit.Ext(up[1])))
				m.Echo("?_t=%d", time.Now().Unix())
			}},
			aaa.THEME: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] != ice.AUTO && tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
					cli.TellApp(m, "System Events", `tell appearance preferences to set dark mode to `+
						kit.Select(ice.TRUE, ice.FALSE, kit.IsIn(kit.Select(html.LIGHT, arg, 0), html.LIGHT, html.WHITE)))
				}
			}},
			aaa.EMAIL: {Name: "email to='shy@shylinux.com' subject content", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.TO) != aaa.UserEmail(m, "") && !aaa.Right(m, aaa.EMAIL, m.Option(nfs.TO)) {
					return
				}
				m.Cmdy(web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN, mdb.TEXT, tcp.PublishLocalhost(m, m.Option(ice.MSG_USERWEB)))
				aaa.SendEmail(m, aaa.ADMIN, m.Option(aaa.TO), "")
			}},
			MESSAGE: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.INPUTS || arg[0] == mdb.ACTION && arg[1] == mdb.INPUTS {
					m.Cmdy(web.Space(m, m.Option(ice.POD)), MESSAGE, arg)
				} else {
					m.Cmdy(web.Space(m, m.Option(ice.POD)), MESSAGE, tcp.SEND, arg).ToastSuccess()
				}
			}},
			aaa.LOGOUT: {Role: aaa.VOID, Hand: aaa.SessLogout},
			web.ONLINE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.STREAM, web.ONLINE) }},
			cli.QRCODE: {Hand: func(m *ice.Message, arg ...string) {
				link := tcp.PublishLocalhost(m, m.OptionDefault(mdb.LINK, m.Option(ice.MSG_USERWEB)))
				m.Push(mdb.NAME, link).PushQRCode(mdb.TEXT, kit.MergeURL(link, ice.FROM_DAEMON, m.Option(ice.MSG_DAEMON)))
			}},
			mdb.CREATE: {Name: "create type*=plugin,qrcode,oauth name* help icons link order space index args"},
			// Hand: func(m *ice.Message, arg ...string) { mdb.HashCreate(m, m.OptionSimple())},

			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { mdb.HashRemove(m, m.OptionSimple(mdb.NAME)) }},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, m.OptionSimple(mdb.NAME), arg) }},
			ice.DEMO: {Help: "体验", Icon: "bi bi-shield-fill-check", Hand: func(m *ice.Message, arg ...string) {
				if kit.IsIn(m.Option(ice.MSG_USERROLE), aaa.TECH, aaa.ROOT) {
					m.Cmd("", mdb.CREATE, mdb.TYPE, mdb.PLUGIN, mdb.NAME, "免登录体验", mdb.ORDER, "12", ctx.INDEX, HEADER, ctx.ARGS, ice.DEMO)
					mdb.Config(m, ice.DEMO, ice.TRUE)
				} else if mdb.Config(m, ice.DEMO) == ice.TRUE {
					web.RenderCookie(m, aaa.SessCreate(m, ice.Info.Username))
					m.Echo("login success")
				} else {
					m.Echo("login failure")
				}
			}},
			ice.INFO: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {}},
			aaa.USER: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, aaa.USER, "msg", http.MethodPost, "/chat/header/action/info", ice.MSG_SESSID, kit.Select(m.Option(ice.MSG_SESSID), arg, 0))
				if msg.Option(ice.MSG_USERNAME) != "" {
					aaa.SessCheck(m, m.Option(ice.MSG_SESSID, aaa.SessCreate(m, msg.Option(ice.MSG_USERNAME))))
					m.Echo(m.Option(ice.MSG_SESSID))
				}
			}},
		}, web.ApiAction(), mdb.ImportantHashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,help,icons,order,link,space,index,args")), Hand: func(m *ice.Message, arg ...string) {
			if kit.Contains(m.Option(ice.MSG_USERUA), "MicroMessenger") {
				if m.Option(ice.MSG_USERNAME) == "" && m.Option("code") == "" && mdb.Config(m, "oauth") != "" {
					m.ProcessOpen(mdb.Config(m, "oauth"))
					// return
				}
			}
			kit.If(kit.Select(m.Option(ice.POD), m.Option(ice.MSG_USERPOD)), func(p string) {
				m.Cmdy(web.SPACE, p, m.PrefixKey(), ice.Maps{ice.MSG_USERPOD: "", ice.POD: ""})
			}, func() {
				m.Option(ice.MSG_NODETYPE, ice.Info.NodeType)
				m.Option(ice.MSG_NODENAME, ice.Info.NodeName)
				m.Option("favicon", ice.Info.NodeIcon)
				if ice.Info.NodeType == web.WORKER && ice.Info.Titles == "ContextOS" {
					return
				}
				m.Option("titles", ice.Info.Titles)
			})
			if ice.Info.NodeType == web.WORKER {
				return
			}
			m.Option(aaa.LANGUAGE, strings.ReplaceAll(strings.ToLower(kit.Select("", kit.Split(kit.GetValid(
				func() string { return kit.Select("", "zh-cn", strings.Contains(m.Option(ice.MSG_USERUA), "zh_CN")) },
				func() string { return kit.Select("", kit.Split(m.R.Header.Get(html.AcceptLanguage), ",;"), 0) },
				func() string { return ice.Info.Lang },
			), " ."), 0)), "_", "-"))
			m.Option("language.list", m.Cmd(nfs.DIR, nfs.TemplatePath(m, aaa.LANGUAGE)+nfs.PS, nfs.FILE).Appendv(nfs.FILE))
			m.Option("theme.list", m.Cmd(nfs.DIR, nfs.TemplatePath(m, aaa.THEME)+nfs.PS, nfs.FILE).Appendv(nfs.FILE))
			if m.Option(ice.MSG_USERNAME) == "" || kit.IsIn(m.Option(ice.MSG_INDEX), m.PrefixKey(), m.CommandKey()) {
				mdb.HashSelect(m, arg...).Sort(mdb.ORDER, ice.INT)
				m.Table(func(value ice.Maps) { m.Push(mdb.STATUS, kit.Select(mdb.ENABLE, mdb.DISABLE, value[mdb.ORDER] == "")) })
				defer m.StatusTimeCount(kit.Dict(mdb.ConfigSimple(m, ice.DEMO)))
			}
			if gdb.Event(m, HEADER_AGENT); !_header_check(m, arg...) {
				kit.If(m.Length() == 0, func() {
					m.Push(mdb.TIME, m.Time()).Push(mdb.NAME, cli.QRCODE).Push(mdb.HELP, "扫码登录").Push(mdb.ICONS, nfs.USR_ICONS_VOLCANOS).Push(mdb.TYPE, cli.QRCODE).Push(web.LINK, "").Push(mdb.ORDER, "10")
				})
				kit.If(GetSSO(m), func(p string) {
					m.Push(mdb.TIME, m.Time()).Push(mdb.NAME, web.SERVE).Push(mdb.ICONS, nfs.USR_ICONS_ICEBERGS).Push(mdb.TYPE, "oauth").Push(web.LINK, p).Push(mdb.ORDER, "100")
				})
			} else {
				kit.If(kit.IsIn(m.Option(ice.MSG_USERROLE), aaa.TECH, aaa.ROOT), func() { m.Action(mdb.CREATE, ice.DEMO) })
				msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
				if role := msg.Append(aaa.USERROLE); role != m.Option(ice.MSG_USERROLE) {
					m.Cmd(aaa.SESS, mdb.MODIFY, mdb.HASH, m.Option(ice.MSG_SESSID), aaa.USERROLE, m.Option(ice.MSG_USERROLE, role))
				}
				kit.For([]string{aaa.USERNICK, aaa.LANGUAGE, aaa.EMAIL}, func(k string) { kit.If(msg.Append(k), func(v string) { m.Option(k, v) }) })
				kit.For([]string{aaa.AVATAR, aaa.BACKGROUND}, func(k string) { m.Option(k, msg.Append(k)) })
			}
		}},
	})
}
func AddHeaderLogin(m *ice.Message, types, name, help, order string, arg ...string) {
	m.Cmd(web.CHAT_HEADER, mdb.CREATE, mdb.TYPE, types, mdb.NAME, name, mdb.HELP, help, mdb.ORDER, order, arg)
}
