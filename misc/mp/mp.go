package mp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	kit "github.com/shylinux/toolkits"
)

const (
	LOGIN = "login"
)
const MP = "mp"

var Index = &ice.Context{Name: MP, Help: "小程序",
	Configs: map[string]*ice.Config{
		LOGIN: {Name: LOGIN, Help: "认证", Value: kit.Data(
			"auth", "/sns/jscode2session?grant_type=authorization_code",
			"weixin", "https://api.weixin.qq.com",
			"appid", "", "appmm", "", "token", "",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(web.SPIDE, mdb.CREATE, "weixin", m.Conf(LOGIN, "meta.weixin"))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},

		"/login/": {Name: "/login/", Help: "登录", Action: map[string]*ice.Action{
			"code": {Name: "code", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, "weixin", web.SPIDE_GET, m.Conf(LOGIN, "meta.auth"), "js_code", m.Option("code"),
					"appid", m.Conf(LOGIN, "meta.appid"), "secret", m.Conf(LOGIN, "meta.appmm"))

				// 用户登录
				m.Option(ice.MSG_USERZONE, MP)
				m.Echo(aaa.SessCreate(msg, msg.Append("openid"), aaa.UserRole(msg, msg.Append("openid"))))
			}},
			"info": {Name: "info", Help: "信息", Hand: func(m *ice.Message, arg ...string) {
				m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
				m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERZONE, MP, aaa.USERNICK, m.Option("nickName"),
					"avatar_url", m.Option("avatarUrl"),
					"gender", kit.Select("女", "男", m.Option("gender") == "1"),
					"country", m.Option("country"), "city", m.Option("city"),
					"language", m.Option("language"),
					"province", m.Option("province"),
				)
			}},
			"scan": {Name: "scan", Help: "scan", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(web.SHARE) != "" {
					if m.Option(chat.RIVER) != "" {
						m.Cmdy(chat.AUTH, mdb.INSERT)
					}
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
