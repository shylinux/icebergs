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
	m.Cmdy(aaa.USER, mdb.MODIFY, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.ActionKey(), m.Option(m.ActionKey(), arg[0]))
}
func _header_share(m *ice.Message, arg ...string) {
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
		web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
			switch kit.Select("", arg, 0) {
			case web.P(HEADER):
				switch kit.Select("", arg, 1) {
				case "", aaa.LOGIN:
					return // 免登录
				}
			default:
				if aaa.Right(m, arg) {
					return // 免登录
				}
			}
			m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, arg)
		}},
		web.P(HEADER): {Name: "/header", Help: "标题栏", Actions: ice.MergeActions(ice.Actions{
			aaa.LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				if aaa.UserLogin(m, arg[0], arg[1]) {
					web.RenderCookie(m, aaa.SessCreate(m, arg[0]))
				}
			}},
			aaa.LOGOUT:     {Hand: func(m *ice.Message, arg ...string) { aaa.UserLogout(m) }},
			aaa.PASSWORD:   {Hand: func(m *ice.Message, arg ...string) { _header_users(m, arg...) }},
			aaa.USERNICK:   {Hand: func(m *ice.Message, arg ...string) { _header_users(m, arg...) }},
			aaa.LANGUAGE:   {Hand: func(m *ice.Message, arg ...string) { _header_users(m, arg...) }},
			aaa.BACKGROUND: {Hand: func(m *ice.Message, arg ...string) { _header_users(m, arg...) }},
			aaa.AVATAR:     {Hand: func(m *ice.Message, arg ...string) { _header_users(m, arg...) }},
			web.SHARE:      {Hand: func(m *ice.Message, arg ...string) { _header_share(m, arg...) }},
			"webpack":      {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("webpack", "build") }},
		}, ctx.ConfAction(aaa.LOGIN, kit.List("密码登录", "扫码授权")), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
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
