package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

func _header_check(m *ice.Message) {
	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) {
		case web.LOGIN: // 扫码登录
			if m.Option(ice.MSG_USERNAME) != msg.Append(aaa.USERNAME) {
				web.RenderCookie(m, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME))))
			}
		}
	}

	if m.Option(ice.MSG_USERNAME) == "" { // 单点登录
		m.Option(web.SSO, m.Conf(web.SERVE, kit.Keym(web.SSO)))
	}
}
func _header_grant(m *ice.Message, arg ...string) {
	if pod := m.Option(kit.SSH_POD); pod != "" {
		m.Option(kit.SSH_POD, "")
		m.Cmd(web.SPACE, pod, m.Prefix(P_HEADER), kit.MDB_ACTION, GRANT, arg)
		return
	}

	m.Cmd(aaa.ROLE, kit.Select(aaa.TECH, aaa.VOID, m.Option(ice.MSG_USERROLE) == aaa.VOID), m.Option(ice.MSG_USERNAME))
	m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
}
func _header_users(m *ice.Message, key string, arg ...string) {
	m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
	m.Cmdy("aaa.user", kit.MDB_ACTION, mdb.MODIFY, key, m.Option(key, arg[0]))
}

const (
	TITLE = "title"
	LOGIN = "login"
	CHECK = "check"
	GRANT = "grant"
	AGENT = "agent"
)
const P_HEADER = "/header"
const HEADER = "header"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			HEADER: {Name: HEADER, Help: "标题栏", Value: kit.Data(TITLE, "github.com/shylinux/contexts")},
		},
		Commands: map[string]*ice.Command{
			P_HEADER: {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
				LOGIN: {Name: "login", Help: "用户登录", Hand: func(m *ice.Message, arg ...string) {
					if aaa.UserLogin(m, arg[0], arg[1]) {
						web.RenderCookie(m, aaa.SessCreate(m, arg[0]))
					}
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},
				CHECK: {Name: "check", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
					_header_check(m)
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},
				GRANT: {Name: "grant space", Help: "用户授权", Hand: func(m *ice.Message, arg ...string) {
					_header_grant(m, arg...)
				}},
				AGENT: {Name: "agent", Help: "应用宿主", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.chat.wx.access", "config")
				}},

				code.WEBPACK: {Name: "webpack", Help: "网页打包", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.WEBPACK, mdb.CREATE)
				}},
				web.SHARE: {Name: "share type", Help: "用户共享", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.SHARE, mdb.CREATE, kit.MDB_TYPE, LOGIN, arg)
				}},
				aaa.AVATAR: {Name: "avatar", Help: "头像图片", Hand: func(m *ice.Message, arg ...string) {
					_header_users(m, aaa.AVATAR, arg...)
				}},
				aaa.BACKGROUND: {Name: "background", Help: "背景图片", Hand: func(m *ice.Message, arg ...string) {
					_header_users(m, aaa.BACKGROUND, arg...)
				}},
				aaa.USERNICK: {Name: "usernick", Help: "用户昵称", Hand: func(m *ice.Message, arg ...string) {
					_header_users(m, aaa.USERNICK, arg...)
				}},
				aaa.USERROLE: {Name: "userrole", Help: "用户角色", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(aaa.UserRole(m, m.Option("who")))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				user := m.Cmd("aaa.user", m.Option(ice.MSG_USERNAME))
				for _, k := range []string{aaa.AVATAR, aaa.BACKGROUND} {
					m.Option(k, user.Append(k))
				}
				m.Echo(m.Conf(HEADER, kit.Keym(TITLE)))
			}},
		},
	})
}
