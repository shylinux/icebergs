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
		msg := m.Cmd(web.SHARE, m.Option(web.SHARE))

		switch msg.Append(kit.MDB_TYPE) {
		case web.LOGIN:
			if m.Option(ice.MSG_USERNAME) != msg.Append(aaa.USERNAME) {
				web.Render(m, web.COOKIE, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME))))
			}
		case web.APPLY:
		}

		m.Option(kit.MDB_TYPE, msg.Append(kit.MDB_TYPE))
		m.Option(kit.MDB_NAME, msg.Append(kit.MDB_NAME))
		m.Option(kit.MDB_TEXT, msg.Append(kit.MDB_TEXT))
		m.Option(aaa.USERNAME, msg.Append(aaa.USERNAME))
	}
	m.Option(web.SSO, m.Conf(web.SERVE, kit.Keym(web.SSO)))
}

const (
	APPLY = "apply"
	LOGIN = "login"
	CHECK = "check"
	TITLE = "title"
	AGENT = "agent"

	BACKGROUND = "background"
)
const P_HEADER = "/header"
const HEADER = "header"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			HEADER: {Name: HEADER, Help: "标题栏", Value: kit.Data(TITLE, "github.com/shylinux/contexts")},
		},
		Commands: map[string]*ice.Command{
			"toast": {Name: "toast target msg", Help: "命令行", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(ice.MSG_USERUA, "chrome")
				m.Cmd(web.SPACE, arg[0], "Header.user.toast", "", ice.Render(m, ice.RENDER_QRCODE, arg[1]), arg[2:])
			}},

			P_HEADER: {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
				"auth": {Name: "auth space", Help: "用户授权", Hand: func(m *ice.Message, arg ...string) {
					if pod := m.Option(kit.SSH_POD); pod != "" {
						m.Option(kit.SSH_POD, "")
						m.Cmd(web.SPACE, pod, m.Prefix(P_HEADER), kit.MDB_ACTION, "auth", arg)
						return
					}

					m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
				}},
				web.SHARE: {Name: "share type", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.SHARE, mdb.CREATE, kit.MDB_TYPE, LOGIN, arg)
				}},
				APPLY: {Name: "apply", Help: "用户登录", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.SHARE, mdb.CREATE, kit.MDB_TYPE, web.APPLY, kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
				LOGIN: {Name: "login", Help: "用户登录", Hand: func(m *ice.Message, arg ...string) {
					if aaa.UserLogin(m, arg[0], arg[1]) {
						web.Render(m, web.COOKIE, aaa.SessCreate(m, arg[0]))
					}
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},
				CHECK: {Name: "check", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
					_header_check(m)
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},
				BACKGROUND: {Name: "background", Help: "背景图片", Hand: func(m *ice.Message, arg ...string) {
					m.Option(BACKGROUND, m.Conf(HEADER, kit.Keym(BACKGROUND), arg[0]))
				}},
				AGENT: {Name: "agent", Help: "宿主机", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.chat.wx.access", "config")
				}},

				code.WEBPACK: {Name: "webpack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.WEBPACK, mdb.CREATE)
				}},
				aaa.USERNICK: {Name: "usernick", Help: "用户昵称", Hand: func(m *ice.Message, arg ...string) {
					m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
					m.Cmdy("aaa.user", kit.MDB_ACTION, mdb.MODIFY, aaa.USERNICK, arg[0])
				}},
				aaa.USERROLE: {Name: "userrole", Help: "用户角色", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(aaa.UserRole(m, m.Option("who")))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(BACKGROUND, m.Conf(HEADER, kit.Keym(BACKGROUND)))
				m.Echo(m.Conf(HEADER, kit.Keym(TITLE)))
			}},
		},
	})
}
