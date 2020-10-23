package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const (
	TITLE = "title"
	LOGIN = "login"
	CHECK = "check"
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
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},

				aaa.USERROLE: {Name: "userrole", Help: "用户角色", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(aaa.UserRole(m, m.Option("who")))
				}},

				"pack": {Name: "pack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.webpack", "create")
				}},
				"wx": {Name: "wx", Help: "微信", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.chat.wx.access", "config")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(m.Conf(HEADER, TITLE))
			}},
		},
	}, nil)
}
