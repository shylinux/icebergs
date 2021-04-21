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
		case web.LOGIN:
			if m.Option(ice.MSG_SESSID) == "" {
				web.Render(m, web.COOKIE, aaa.SessCreate(m, msg.Append(aaa.USERNAME)))
			}
		}
	}
	m.Option(web.SSO, m.Conf(web.SERVE, kit.Keym(web.SSO)))
}

const (
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
			P_HEADER: {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
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
