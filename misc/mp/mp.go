package mp

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"

	"net/http"
	"path"
)

var Index = &ice.Context{Name: "mp", Help: "小程序",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"login": {Name: "login", Help: "认证", Value: kit.Data(
			"auth", "/sns/jscode2session?grant_type=authorization_code",
			"weixin", "https://api.weixin.qq.com",
			"appid", "", "appmm", "", "token", "",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(web.SPIDE, mdb.CREATE, "weixin", m.Conf("login", "meta.weixin"))
			m.Confm("login", "meta.userrole", func(key string, value string) {
				m.Cmd(aaa.ROLE, value, key)
			})
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("login")
		}},

		"/login/": {Name: "/login/", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "code":
				msg := m.Cmd(web.SPIDE, "weixin", http.MethodGet, m.Conf("login", "meta.auth"), "js_code", m.Option("code"),
					"appid", m.Conf("login", "meta.appid"), "secret", m.Conf("login", "meta.appmm"))

				// 用户登录
				m.Echo(aaa.SessCreate(msg, msg.Append("openid"), aaa.UserRole(msg, msg.Append("openid"))))

			case "info":
				// 用户信息
				m.Cmd(aaa.USER, mdb.MODIFY, m.Option(ice.MSG_USERNAME),
					"gender", m.Option("gender"),
					"avatar", m.Option("avatarUrl"),
					"usernick", m.Option("nickName"),
					"language", m.Option("language"),
					"province", m.Option("province"),
					"country", m.Option("country"),
					"city", m.Option("city"),
				)

			case "scan":
				if p := m.Option("pod"); p != "" {
					m.Option("pod", "")
					// 代理列表
					m.Cmdy(web.SPACE, p, "web.chat.qrcode.qrcode", mdb.INSERT, "qrcode", m.Option("name"), m.Option("text"))
					return
				}
				m.Cmdy("web.chat.qrcode.qrcode", mdb.INSERT, "qrcode", m.Option("name"), m.Option("text"))

			case "auth":
				if !m.Options(ice.MSG_USERNAME) {
					m.Render("status", 401, "not login")
					break
				}

				switch kit.Select("active", m.Option("type")) {
				case "share":
					m.Richs(web.SHARE, nil, m.Option("text"), func(key string, value map[string]interface{}) {
						switch value["type"] {
						case "invite":
							if m.Option(ice.MSG_USERROLE) != value["name"] {
								m.Cmd(aaa.ROLE, value["name"], m.Option(ice.MSG_USERNAME))
								m.Cmd("web.chat.auto", m.Option(ice.MSG_USERNAME), value["name"])
							}
							break
						default:
							m.Option("type", value["type"])
							m.Option("name", value["name"])
							m.Option("text", value["text"])
						}
					})
				}

				switch kit.Select("active", m.Option("type")) {
				case "active":
					// 网页登录
					m.Cmdy(web.SPACE, m.Option("name"), "sessid", m.Cmdx(aaa.SESS, "create", m.Option(ice.MSG_USERNAME)))
				case "login":
					// 终端登录
					m.Cmdy(aaa.SESS, "auth", m.Option("text"), m.Option(ice.MSG_USERNAME))
				}

			case "upload":
				msg := m.Cmd(web.CACHE, "upload")
				m.Cmd(web.STORY, web.WATCH, msg.Append("data"), path.Join("usr/local/mp/", path.Base(msg.Append("name"))))
				m.Cmd(web.FAVOR, "device", "file", msg.Append("name"), msg.Append("data"))
				m.Render(msg.Append("data"))

			case "cmds":
				if !m.Options(ice.MSG_USERNAME) {
					m.Render("status", 401, "not login")
					break
				}
				if arg = kit.Split(arg[1]); !m.Right(arg) {
					m.Render("status", 403, "not auth")
					break
				}

				// 执行命令
				m.Cmdy(arg)
			}
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
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
