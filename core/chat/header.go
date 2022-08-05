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
	kit "shylinux.com/x/toolkits"
)

func _header_check(m *ice.Message, arg ...string) bool {
	if m.Option(web.SHARE) != "" {
		m.Cmd(web.SHARE, m.Option(web.SHARE), ice.OptionFields(""), func(value ice.Maps) {
			switch value[mdb.TYPE] {
			case web.FIELD, web.STORM:
				m.Option(ice.MSG_USERNAME, value[aaa.USERNAME])
				m.Option(ice.MSG_USERROLE, value[aaa.USERROLE])
			}
		})
	}
	if m.Option(ice.MSG_USERNAME) != "" {
		return true
	}

	m.Option(web.SSO, m.Config(web.SSO))
	m.Option(web.LOGIN, m.Config(web.LOGIN))
	if m.Option("login.dev", m.CmdAppend(web.SPACE, ice.DEV, mdb.TEXT)) == "" {
		m.Option("login.dev", m.CmdAppend(web.SPACE, ice.SHY, mdb.TEXT))
	}
	return false
}
func _header_share(m *ice.Message, arg ...string) {
	if m.Option(mdb.LINK) == "" {
		m.Cmdy(web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN, arg)
	} else {
		m.Option(mdb.LINK, tcp.ReplaceLocalhost(m, m.Option(mdb.LINK)))
	}
	m.Option(mdb.LINK, kit.MergeURL(m.Option(mdb.LINK), RIVER, "", STORM, ""))
	m.PushQRCode(mdb.TEXT, m.Option(mdb.LINK))
	m.Push(mdb.NAME, m.Option(mdb.LINK))
}
func _header_users(m *ice.Message, key string, arg ...string) {
	m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
	m.Cmdy(aaa.USER, mdb.MODIFY, key, m.Option(key, arg[0]))
}

const (
	TITLE = "title"
	MENUS = "menus"
	TRANS = "trans"

	HEADER_AGENT = "header.agent"
)
const HEADER = "header"

func init() {
	Index.MergeCommands(ice.Commands{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
			m.Debug("what %v", m.FormatStack(1, 100))
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
		HEADER: {Name: "header", Help: "标题栏", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.CommandKey())
			}},
			aaa.LOGIN: {Name: "login", Help: "密码登录", Hand: func(m *ice.Message, arg ...string) {
				if aaa.UserLogin(m, arg[0], arg[1]) {
					web.RenderCookie(m, aaa.SessCreate(m, arg[0]))
				}
			}},
			aaa.LOGOUT: {Name: "logout", Help: "退出登录", Hand: func(m *ice.Message, arg ...string) {
				aaa.UserLogout(m)
			}},
			aaa.PASSWORD: {Name: "password", Help: "修改密码", Hand: func(m *ice.Message, arg ...string) {
				_header_users(m, m.ActionKey(), arg...)
			}},
			aaa.USERNICK: {Name: "usernick", Help: "用户昵称", Hand: func(m *ice.Message, arg ...string) {
				_header_users(m, m.ActionKey(), arg...)
			}},
			aaa.AVATAR: {Name: "avatar", Help: "用户头像", Hand: func(m *ice.Message, arg ...string) {
				_header_users(m, m.ActionKey(), arg...)
			}},
			aaa.BACKGROUND: {Name: "background", Help: "用户壁纸", Hand: func(m *ice.Message, arg ...string) {
				_header_users(m, m.ActionKey(), arg...)
			}},
			aaa.LANGUAGE: {Name: "language", Help: "语言地区", Hand: func(m *ice.Message, arg ...string) {
				_header_users(m, m.ActionKey(), arg...)
			}},
			ctx.CONFIG: {Name: "config scope", Help: "拉取配置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPACE, m.Option(ice.MSG_USERPOD), m.Prefix("oauth.oauth"), "check", arg)
			}},
			"webpack": {Name: "webpack", Help: "打包页面", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("webpack", cli.BUILD, m.OptionSimple(mdb.NAME))
			}},
			web.SHARE: {Name: "share type", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_header_share(m, arg...)
			}},
		}, ctx.ConfAction(aaa.LOGIN, kit.List("登录", "扫码")), web.ApiAction()), Hand: func(m *ice.Message, arg ...string) {
			if !_header_check(m, arg...) {
				return
			}

			msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
			for _, k := range []string{aaa.USERNICK, aaa.LANGUAGE} {
				m.Option(k, msg.Append(k))
			}
			for _, k := range []string{aaa.AVATAR, aaa.BACKGROUND} {
				if msg.Append(k) != "" && aaa.Right(m.Spawn(), msg.Append(k)) {
					m.Option(k, web.SHARE_LOCAL+k)
				}
			}
			if m.Option(aaa.AVATAR) == "" && m.R.Header.Get("Staffname") != "" {
				m.Option(aaa.AVATAR, kit.Format("https://dayu.oa.com/avatars/%s/profile.jpg", m.R.Header.Get("Staffname")))
			}

			gdb.Event(m, HEADER_AGENT)
			m.Option(TRANS, kit.Format(kit.Value(m.Target().Commands[web.P(m.CommandKey())].Meta, "_trans")))
			m.Option(MENUS, m.Config(MENUS))
			m.Echo(m.Config(TITLE))
		}},
	})
}
