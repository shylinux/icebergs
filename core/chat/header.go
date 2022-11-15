package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
)

func _header_users(m *ice.Message, arg ...string) {
	if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
		return
	}
	if m.Warn(m.Option(web.SHARE) != "", ice.ErrNotRight, "没有权限") {
		return
	}
	m.Cmdy(aaa.USER, mdb.MODIFY, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.ActionKey(), m.Option(m.ActionKey(), arg[0]))
}
func _header_share(m *ice.Message, arg ...string) {
	if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, "没有登录") {
		return
	}
	if m.Warn(m.Option(web.SHARE) != "", ice.ErrNotRight, "没有权限") {
		return
	}
	for i := 0; i < len(arg)-1; i += 2 {
		m.Option(arg[i], arg[i+1])
	}
	if m.Option(mdb.LINK) == "" {
		m.Cmdy(web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN, arg)
	} else {
		m.Option(mdb.LINK, tcp.ReplaceLocalhost(m, m.Option(mdb.LINK)))
	}
	m.Push(mdb.NAME, m.Option(mdb.LINK)).PushQRCode(mdb.TEXT, m.Option(mdb.LINK))
}
func _header_check(m *ice.Message, arg ...string) bool {
	if m.Option(web.SHARE) != "" {
		m.Cmd(web.SHARE, m.Option(web.SHARE), ice.OptionFields(""), func(value ice.Maps) {
			if web.IsNotValidShare(m, value) {
				return
			}
			switch value[mdb.TYPE] {
			case web.STORM, web.FIELD:
				aaa.SessAuth(m, value)
			}
		})
	}
	if m.Option(ice.MSG_USERNAME) != "" {
		return true
	}
	if m.OptionFromConfig(web.SSO) == "" && m.Option(ice.DEV, m.CmdAppend(web.SPACE, ice.DEV, mdb.TEXT)) == "" {
		if m.Option(ice.DEV, m.CmdAppend(web.SPACE, ice.SHY, mdb.TEXT)) == "" {
			m.OptionFromConfig(web.LOGIN)
		}
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
			aaa.LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 1 && aaa.UserLogin(m, arg[0], arg[1]) {
					web.RenderCookie(m, aaa.SessCreate(m, arg[0]))
				}
			}},
			aaa.LOGOUT:     {Hand: aaa.UserLogout},
			aaa.PASSWORD:   {Hand: _header_users},
			aaa.USERNICK:   {Hand: _header_users},
			aaa.LANGUAGE:   {Hand: _header_users},
			aaa.BACKGROUND: {Hand: _header_users},
			aaa.AVATAR:     {Hand: _header_users},
			web.SHARE:      {Hand: _header_share},
			"webpack":      {Hand: ctx.CmdHandler("webpack", "build")},
		}, aaa.BlackAction("webpack")), Hand: func(m *ice.Message, arg ...string) {
			if gdb.Event(m, HEADER_AGENT); !_header_check(m, arg...) {
				return
			}
			msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
			for _, k := range []string{aaa.USERNICK, aaa.LANGUAGE} {
				m.Option(k, msg.Append(k))
			}
			for _, k := range []string{aaa.BACKGROUND, aaa.AVATAR} {
				if strings.HasPrefix(msg.Append(k), ice.HTTP) {
					m.Option(k, msg.Append(k))
				} else if msg.Append(k) != "" && aaa.Right(m.Spawn(), msg.Append(k)) {
					m.Option(k, web.SHARE_LOCAL+k)
				}
			}
			m.Echo(m.Config(TITLE)).OptionFromConfig(MENUS)
		}},
	})
}
