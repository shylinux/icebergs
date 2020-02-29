package mp

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "mp", Help: "小程序",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"login": {Name: "login", Help: "认证", Value: kit.Data(
			"auth", "/sns/jscode2session?grant_type=authorization_code",
			"weixin", "https://api.weixin.qq.com",
			"appid", "", "appmm", "", "token", "",
			"userrole", kit.Dict(),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Confm("login", "meta.userrole", func(key string, value string) {
				m.Cmd(ice.AAA_ROLE, value, key)
			})
			m.Cmd(ice.WEB_SPIDE, "add", "weixin", m.Conf("login", "meta.weixin"))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("login")
		}},
		"/login/": {Name: "/login/", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "code":
				msg := m.Cmd(ice.WEB_SPIDE, "weixin", "GET", m.Conf("login", "meta.auth"),
					"js_code", m.Option("code"), "appid", m.Conf("login", "meta.appid"), "secret", m.Conf("login", "meta.appmm"))

				// 用户登录
				m.Option(ice.MSG_USERNAME, msg.Append("openid"))
				m.Option(ice.MSG_USERROLE, m.Cmdx(ice.AAA_ROLE, "check", m.Option("FromUserName")))
				m.Info("%s: %s", m.Option(ice.MSG_USERROLE), m.Option(ice.MSG_USERNAME))
				m.Echo(m.Option(ice.MSG_SESSID, m.Cmdx(ice.AAA_USER, "login", m.Option(ice.MSG_USERNAME))))

			case "info":
				// 用户信息
				m.Richs(ice.AAA_SESS, nil, m.Option(ice.MSG_SESSID), func(key string, value map[string]interface{}) {
					m.Richs(ice.AAA_USER, nil, value["username"], func(key string, value map[string]interface{}) {
						value["gender"] = m.Option("gender")
						value["avatar"] = m.Option("avatarUrl")
						value["nickname"] = m.Option("nickName")
						value["language"] = m.Option("language")
						value["province"] = m.Option("province")
						value["country"] = m.Option("country")
						value["city"] = m.Option("city")
					})
				})

			case "scan":
				m.Echo(m.Option("scan")).Push("_output", "qrcode")

			case "auth":
				if !m.Options(ice.MSG_USERNAME) || !m.Options(ice.MSG_SESSID) {
					m.Echo("401").Push("_output", "status")
					break
				}
				switch kit.Select("active", m.Option("type")) {
				case "active":
					// 授权登录
					m.Cmd(ice.WEB_SPACE, "auth", m.Option("auth"), m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERROLE))
				}

			case "cmds":
				if arg = kit.Split(arg[1]); !m.Right(arg) {
					return
				}

				// 执行命令
				msg := m.Cmd(arg)
				if m.Hand = false; !msg.Hand {
					msg = m.Cmd(ice.CLI_SYSTEM, arg)
				}
			}
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
