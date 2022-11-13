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
	kit "shylinux.com/x/toolkits"
)

func _header_users(m *ice.Message, arg ...string) {
	if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
		return
	}
	m.Cmdy(aaa.USER, mdb.MODIFY, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.ActionKey(), m.Option(m.ActionKey(), arg[0]))
}
func _header_share(m *ice.Message, arg ...string) {
	if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
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
			if m.Warn(value[mdb.TIME] < m.Time(), ice.ErrNotValid, m.Option(web.SHARE), value[mdb.TIME], m.Time()) {
				return
			}
			switch value[mdb.TYPE] {
			case web.LOGIN:
				if value[aaa.USERNAME] != m.Option(ice.MSG_USERNAME) {
					web.RenderCookie(m, aaa.SessCreate(m, value[aaa.USERNAME]))
				}
				fallthrough
			case web.STORM, web.FIELD:
				aaa.SessAuth(m, value)
			}
		})
	}
	if m.Option(ice.MSG_USERNAME) != "" {
		return true
	}
	if m.OptionFromConfig(web.SSO) == "" && m.Option("login.dev", m.CmdAppend(web.SPACE, ice.DEV, mdb.TEXT)) == "" {
		if m.Option("login.dev", m.CmdAppend(web.SPACE, ice.SHY, mdb.TEXT)) == "" {
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
		}, ctx.ConfAction(aaa.LOGIN, kit.List("密码登录", "扫码授权")), aaa.BlackAction("webpack")), Hand: func(m *ice.Message, arg ...string) {
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
