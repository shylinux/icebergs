package mp

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"
	"time"
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
			m.Cmd(ice.CTX_CONFIG, "load", "mp.json")
			m.Confm("login", "meta.userrole", func(key string, value string) {
				m.Cmd(ice.AAA_ROLE, value, key)
			})
			m.Cmd(ice.WEB_SPIDE, "add", "weixin", m.Conf("login", "meta.weixin"))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "mp.json", "web.chat.mp.login")
		}},
		"/login/": {Name: "/login/", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "code":
				msg := m.Cmd(ice.WEB_SPIDE, "weixin", "GET", m.Conf("login", "meta.auth"),
					"js_code", m.Option("code"), "appid", m.Conf("login", "meta.appid"), "secret", m.Conf("login", "meta.appmm"))

				if m.Richs(ice.AAA_USER, nil, msg.Append("openid"), nil) == nil {
					// 创建用户
					m.Rich(ice.AAA_USER, nil, kit.Dict(
						"username", msg.Append("openid"),
						"expires_in", time.Now().Unix()+kit.Int64(msg.Append("expires_in")),
						"session_key", msg.Append("session_key"),
						"usernode", m.Conf(ice.CLI_RUNTIME, "boot.hostname"),
					))
					m.Event(ice.USER_CREATE, msg.Append("openid"))
				}

				if m.Options(ice.MSG_SESSID) && m.Cmdx(ice.AAA_SESS, "check", m.Option(ice.MSG_SESSID)) == msg.Append("openid") {
					// 复用会话
					m.Echo(m.Option(ice.MSG_SESSID))
				} else {
					// 创建会话
					role := m.Conf("login", kit.Keys("meta.userrole", msg.Append("openid")))
					sessid := m.Rich(ice.AAA_SESS, nil, kit.Dict(
						"username", msg.Append("openid"), "userrole", role,
					))
					m.Info("user: %s role: %s sess: %s", msg.Append("openid"), role, sessid)
					m.Echo(msg.Option(ice.MSG_SESSID, sessid))
				}

			case "info":
				m.Richs(ice.AAA_SESS, nil, m.Option(ice.MSG_SESSID), func(key string, value map[string]interface{}) {
					m.Richs(ice.AAA_USER, nil, value["username"], func(key string, value map[string]interface{}) {
						// 注册用户
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
				m.Log("fuck", "what %v", m.Option(ice.MSG_USERNAME))
				m.Log("fuck", "what %v", m.Option(ice.MSG_SESSID))
				if !m.Options(ice.MSG_USERNAME) || !m.Options(ice.MSG_SESSID) {
					m.Echo("401").Push("_output", "status")
					break
				}
				m.Cmd(ice.WEB_SPACE, "auth", m.Option("auth"), m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERROLE))

			case "cmds":
				if arg = kit.Split(arg[1]); m.Right(arg) {
					m.Hand = false
					if m.Cmdy(arg); !m.Hand {
						m.Set("result")
						m.Cmdy(ice.CLI_SYSTEM, arg)
					}
				}
			}
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
