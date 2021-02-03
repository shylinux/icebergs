package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const (
	TITLE = "title"
	LOGIN = "login"
	CHECK = "check"

	BACKGROUND = "background"
)
const HEADER = "header"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			HEADER: {Name: HEADER, Help: "标题栏", Value: kit.Dict(TITLE, "github.com/shylinux/contexts")},
		},
		Commands: map[string]*ice.Command{
			"/header": {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
				LOGIN: {Name: "login", Help: "用户登录", Hand: func(m *ice.Message, arg ...string) {
					if aaa.UserLogin(m, arg[0], arg[1]) {
						m.Option(ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERROLE)))
						web.Render(m, web.COOKIE, m.Option(ice.MSG_SESSID))
					}
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},
				CHECK: {Name: "check", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(web.SHARE) != "" {
						switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) {
						case web.LOGIN:
							if m.Option(ice.MSG_SESSID) == "" {
								m.Option(ice.MSG_SESSID, aaa.SessCreate(m, msg.Append(aaa.USERNAME), msg.Append(aaa.USERROLE)))
								web.Render(m, web.COOKIE, m.Option(ice.MSG_SESSID))
							}
						}
					}

					m.Option("sso", m.Conf(web.SERVE, "meta.sso"))
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},

				aaa.USERROLE: {Name: "userrole", Help: "用户角色", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(aaa.UserRole(m, m.Option("who")))
				}},
				BACKGROUND: {Name: "background", Help: "背景图片", Hand: func(m *ice.Message, arg ...string) {
					m.Option(BACKGROUND, m.Conf(HEADER, BACKGROUND, arg[0]))
				}},
				code.WEBPACK: {Name: "webpack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.WEBPACK, mdb.CREATE)
				}},
				"wx": {Name: "wx", Help: "微信", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.chat.wx.access", "config")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(BACKGROUND, m.Conf(HEADER, BACKGROUND))
				m.Echo(m.Conf(HEADER, TITLE))
			}},
		},
	})
}
