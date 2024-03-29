package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
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
	}
	m.Cmdy(aaa.USER, mdb.MODIFY, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.ActionKey(), m.Option(m.ActionKey(), arg[0]))
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
	m.Option(ice.MAIN, mdb.Conf(m, "web.serve", "meta.main"))
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
			case web.STORM, web.FIELD:
				aaa.SessAuth(m, kit.Dict(value))
			}
		})
	}
	if m.Option(ice.MSG_USERNAME) != "" {
		return true
	} else if ctx.OptionFromConfig(m, SSO) == "" && ctx.OptionFromConfig(m, web.LOGIN) == "" {
		m.Option(SSO, GetSSO(m))
	}
	return false
}

const (
	TITLE = "title"
	MENUS = "menus"

	HEADER_AGENT = "header.agent"
)
const HEADER = "header"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(HEADER): {Name: "/header", Help: "标题栏", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT:   {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, HEADER) }},
			mdb.INPUTS:     {Hand: func(m *ice.Message, arg ...string) {}},
			aaa.LOGIN:      {Hand: func(m *ice.Message, arg ...string) {}},
			aaa.LOGOUT:     {Hand: aaa.SessLogout},
			aaa.PASSWORD:   {Hand: _header_users},
			aaa.USERNICK:   {Hand: _header_users},
			aaa.LANGUAGE:   {Hand: _header_users},
			aaa.BACKGROUND: {Hand: _header_users},
			aaa.AVATAR:     {Hand: _header_users},
			web.SHARE:      {Hand: _header_share},
			"webpack":      {Hand: ctx.CmdHandler("webpack", "build")},
			"theme": {Hand: func(m *ice.Message, arg ...string) {
				if !tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
					return
				}
				m.Cmd(cli.SYSTEM, "osascript", "-e", `tell app "System Events" to tell appearance preferences to set dark mode to `+
					kit.Select(ice.FALSE, ice.TRUE, kit.IsIn(kit.Select(html.DARK, arg, 0), html.DARK, html.BLACK)))
			}},
		}, ctx.ConfAction(SSO, "", aaa.LANGUAGE, "zh")), Hand: func(m *ice.Message, arg ...string) {
			if gdb.Event(m, HEADER_AGENT); !_header_check(m, arg...) {
				return
			}
			msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
			kit.For([]string{aaa.USERNICK, aaa.LANGUAGE}, func(k string) { m.Option(k, msg.Append(k)) })
			kit.For([]string{aaa.AVATAR, aaa.BACKGROUND}, func(k string) { m.Option(k, web.RequireFile(m, msg.Append(k))) })
			kit.If(m.Option(aaa.LANGUAGE) == "", func() { m.Option(aaa.LANGUAGE, mdb.Config(m, aaa.LANGUAGE)) })
			m.Echo(mdb.Config(m, TITLE)).Option(MENUS, mdb.Config(m, MENUS))
		}},
	})
}
