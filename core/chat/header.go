package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

func _header_check(m *ice.Message, arg ...string) {
	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) {
		case web.LOGIN: // 扫码登录
			if m.Option(ice.MSG_USERNAME) != msg.Append(aaa.USERNAME) {
				web.RenderCookie(m, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME))))
			}
		case web.STORM:
			m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME))
		case web.FIELD:
			m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME))
		}
	}

	if m.Option(ice.MSG_USERNAME) == "" { // 单点登录
		m.Option(web.SSO, m.Conf(web.SERVE, kit.Keym(web.SSO)))
	}
}
func _header_share(m *ice.Message, arg ...string) {
	if m.Option(kit.MDB_LINK) == "" {
		share := m.Cmdx(web.SHARE, mdb.CREATE, kit.MDB_TYPE, web.LOGIN, arg)
		m.Option(kit.MDB_LINK, kit.MergeURL(m.Option(ice.MSG_USERWEB), web.SHARE, share))
	}
	link := tcp.ReplaceLocalhost(m, m.Option(kit.MDB_LINK))

	m.Set(kit.MDB_NAME, kit.MDB_TEXT)
	m.PushQRCode(kit.MDB_TEXT, link)
	m.Push(kit.MDB_NAME, link)
}
func _header_grant(m *ice.Message, arg ...string) {
	if m.PodCmd(m.Prefix("/header"), ctx.ACTION, GRANT, arg) {
		return // 下发命令
	}

	// 授权登录
	m.Cmd(aaa.ROLE, kit.Select(aaa.TECH, aaa.VOID, m.Option(ice.MSG_USERROLE) == aaa.VOID), m.Option(ice.MSG_USERNAME))
	m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
}
func _header_users(m *ice.Message, key string, arg ...string) {
	m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
	m.Cmdy("aaa.user", ctx.ACTION, mdb.MODIFY, key, m.Option(key, arg[0]))
}

const (
	TITLE = "title"
	AGENT = "agent"
	CHECK = "check"
	SHARE = "share"
	GRANT = "grant"
)
const HEADER = "header"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			HEADER: {Name: HEADER, Help: "标题栏", Value: kit.Data(TITLE, "github.com/shylinux/contexts")},
		},
		Commands: map[string]*ice.Command{
			"/header": {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
				AGENT: {Name: "agent", Help: "应用宿主", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.chat.wx.access", "config")
				}},
				CHECK: {Name: "check", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
					_header_check(m, arg...)
				}},
				SHARE: {Name: "share type", Help: "扫码登录", Hand: func(m *ice.Message, arg ...string) {
					_header_share(m, arg...)
				}},
				GRANT: {Name: "grant space", Help: "扫码授权", Hand: func(m *ice.Message, arg ...string) {
					_header_grant(m, arg...)
				}},

				aaa.LOGIN: {Name: "login", Help: "密码登录", Hand: func(m *ice.Message, arg ...string) {
					if aaa.UserLogin(m, arg[0], arg[1]) {
						web.RenderCookie(m, aaa.SessCreate(m, arg[0]))
					}
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},
				aaa.LOGOUT: {Name: "logout", Help: "退出登录", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(aaa.SESS, mdb.REMOVE, ice.OptionHash(m.Option(ice.MSG_SESSID)))
				}},

				aaa.AVATAR: {Name: "avatar", Help: "头像图片", Hand: func(m *ice.Message, arg ...string) {
					_header_users(m, aaa.AVATAR, arg...)
				}},
				aaa.USERNICK: {Name: "usernick", Help: "用户昵称", Hand: func(m *ice.Message, arg ...string) {
					_header_users(m, aaa.USERNICK, arg...)
				}},
				aaa.BACKGROUND: {Name: "background", Help: "背景图片", Hand: func(m *ice.Message, arg ...string) {
					_header_users(m, aaa.BACKGROUND, arg...)
				}},
				code.WEBPACK: {Name: "webpack", Help: "网页打包", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.WEBPACK, mdb.CREATE, m.OptionSimple(kit.MDB_NAME))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				user := m.Cmd("aaa.user", m.Option(ice.MSG_USERNAME))
				for _, k := range []string{aaa.BACKGROUND, aaa.AVATAR} {
					m.Option(k, user.Append(k))
				}
				m.Echo(m.Conf(HEADER, kit.Keym(TITLE)))
			}},
		},
	})
}
