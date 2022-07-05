package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _header_agent(m *ice.Message, arg ...string) {
	if strings.Index(m.Option(ice.MSG_USERUA), "MicroMessenger") > -1 {
		m.Cmdy("web.chat.wx.access", "config")
	}
}
func _header_check(m *ice.Message, arg ...string) bool {
	if m.Option(ice.MSG_USERNAME) != "" {
		return true
	}

	m.Option(web.LOGIN, m.Config(web.LOGIN))
	m.Option(web.SSO, m.Conf(web.SERVE, kit.Keym(web.SSO)))
	if m.Option("login.dev", m.Cmd(web.SPACE, ice.DEV).Append(mdb.TEXT)) == "" {
		m.Option("login.dev", m.Cmd(web.SPACE, ice.SHY).Append(mdb.TEXT))
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
	m.Cmdy(aaa.USER, ctx.ACTION, mdb.MODIFY, key, m.Option(key, arg[0]))
}

const (
	TITLE = "title"
	TOPIC = "topic"
	MENUS = "menus"
	TRANS = "trans"
	AGENT = "agent"
	CHECK = "check"
	SHARE = "share"
)
const HEADER = "header"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		HEADER: {Name: HEADER, Help: "标题栏", Value: kit.Data(aaa.LOGIN, kit.List("登录", "扫码"))},
	}, Commands: map[string]*ice.Command{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
			switch arg[0] {
			case "/header":
				if kit.Select("", arg, 1) == "" {
					return // 免登录
				}
			case "/pod/", "/cmd/":
				return // 免登录
			case "/sso":
				return // 免登录
			}
			m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, arg)
		}},
		web.P(HEADER): {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
			AGENT: {Name: "agent", Help: "宿主应用", Hand: func(m *ice.Message, arg ...string) {
				_header_agent(m, arg...)
			}},
			CHECK: {Name: "check", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
				_header_check(m, arg...)
			}},
			SHARE: {Name: "share type", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_header_share(m, arg...)
			}},
			aaa.LOGIN: {Name: "login", Help: "密码登录", Hand: func(m *ice.Message, arg ...string) {
				if aaa.UserLogin(m, arg[0], arg[1]) {
					web.RenderCookie(m, aaa.SessCreate(m, arg[0]))
				}
			}},
			aaa.LOGOUT: {Name: "logout", Help: "退出登录", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.SESS, mdb.REMOVE, kit.Dict(mdb.HASH, m.Option(ice.MSG_SESSID)))
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
				m.Cmdy(web.SPACE, m.Option(ice.MSG_USERPOD), m.Prefix("oauth.oauth"), CHECK, arg)
			}},
			code.WEBPACK: {Name: "webpack", Help: "打包页面", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.WEBPACK, cli.BUILD, m.OptionSimple(mdb.NAME))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if !_header_check(m, arg...) {
				return
			}
			_header_agent(m, arg...)

			msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
			for _, k := range []string{aaa.USERNICK, aaa.LANGUAGE} {
				m.Option(k, msg.Append(k))
			}
			for _, k := range []string{aaa.AVATAR, aaa.BACKGROUND} {
				m.Option(k, kit.Select(web.SHARE_LOCAL+k, kit.Select("void", msg.Append(k)), m.Right(msg.Append(k))))
			}
			if m.Option(aaa.AVATAR) == "" && m.R.Header.Get("Staffname") != "" {
				m.Option(aaa.AVATAR, kit.Format("https://dayu.oa.com/avatars/%s/profile.jpg", m.R.Header.Get("Staffname")))
			}

			m.Option(TRANS, kit.Format(kit.Value(m.Target().Commands[web.P(m.CommandKey())].Meta, "_trans")))
			m.Option(MENUS, m.Config(MENUS))
			m.Echo(m.Config(TITLE))
			// m.Cmdy(WEBSITE)
		}},
	}})
}
