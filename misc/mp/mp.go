package mp

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"

	"path"
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
			m.Cmd(ice.WEB_SPIDE, "add", "weixin", m.Conf("login", "meta.weixin"))
			m.Confm("login", "meta.userrole", func(key string, value string) {
				m.Cmd(ice.AAA_ROLE, value, key)
			})
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("login")
		}},

		"scan": {Name: "scan", Help: "扫码", List: kit.List(
			kit.MDB_INPUT, "text", "name", "location", "cb", "location",
			kit.MDB_INPUT, "text", "name", "battery", "cb", "battery",
			kit.MDB_INPUT, "text", "name", "paste", "cb", "paste",
			kit.MDB_INPUT, "text", "name", "scan", "cb", "scan",
			kit.MDB_INPUT, "text", "name", "wifi", "cb", "wifi",

			kit.MDB_INPUT, "text", "name", "album", "cb", "album",
			kit.MDB_INPUT, "text", "name", "finger", "cb", "finger",
			kit.MDB_INPUT, "text", "name", "vibrate", "cb", "vibrate",
			kit.MDB_INPUT, "text", "name", "wifiList", "cb", "wifiList",
			kit.MDB_INPUT, "text", "name", "wifiConn", "cb", "wifiConn",

			kit.MDB_INPUT, "textarea", "name", "scan", "cb", "scan",
			kit.MDB_INPUT, "textarea", "name", "location", "cb", "location",
			kit.MDB_INPUT, "button", "name", "scan", "cb", "scan",
			kit.MDB_INPUT, "button", "name", "location", "cb", "location",
			kit.MDB_INPUT, "button", "name", "text",
			kit.MDB_INPUT, "button", "name", "share",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(arg[0])
		}},
		"/login/": {Name: "/login/", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "code":
				msg := m.Cmd(ice.WEB_SPIDE, "weixin", "GET", m.Conf("login", "meta.auth"), "js_code", m.Option("code"),
					"appid", m.Conf("login", "meta.appid"), "secret", m.Conf("login", "meta.appmm"))

				// 用户登录
				m.Echo(m.Option(ice.MSG_SESSID, m.Cmdx(ice.AAA_USER, "login", msg.Append("openid"))))

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
				web.Render(m, "qrcode", m.Option("scan"))

			case "auth":
				if !m.Options(ice.MSG_USERNAME) {
					web.Render(m, "status", 401, "not login")
					break
				}

				switch kit.Select("active", m.Option("type")) {
				case "active":
					// 授权登录
					m.Cmd(ice.WEB_SPACE, m.Option("auth"), "sessid", m.Cmdx(ice.AAA_SESS, "create", m.Option(ice.MSG_USERNAME)))
				}

			case "upload":
				msg := m.Cmd(ice.WEB_CACHE, "upload")
				m.Cmdy(ice.WEB_STORY, ice.STORY_WATCH, msg.Append("data"), path.Join("usr/local/mp/", path.Base(msg.Append("name"))))
				web.Render(m, msg.Append("data"))
				m.Cmdy(ice.WEB_FAVOR, "device", "file", msg.Append("name"), msg.Append("data"))

			case "cmds":
				if !m.Options(ice.MSG_USERNAME) {
					web.Render(m, "status", 401, "not login")
					break
				}
				if arg = kit.Split(arg[1]); !m.Right(arg) {
					web.Render(m, "status", 403, "not auth")
					break
				}

				// 执行命令
				m.Cmdy(arg)
			}
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
