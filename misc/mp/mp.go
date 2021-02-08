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
	LOGIN  = "login"
	APPID  = "appid"
	APPMM  = "appmm"
	ACCESS = "access"
	OPENID = "openid"
	WEIXIN = "weixin"
)
const MP = "mp"

var Index = &ice.Context{Name: MP, Help: "小程序",
	Configs: map[string]*ice.Config{
		LOGIN: {Name: LOGIN, Help: "认证", Value: kit.Data(
			WEIXIN, "https://api.weixin.qq.com", APPID, "", APPMM, "",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(web.SPIDE, mdb.CREATE, WEIXIN, m.Conf(LOGIN, kit.Keys(kit.MDB_META, WEIXIN)))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},

		ACCESS: {Name: "access appid auto login", Help: "认证", Action: map[string]*ice.Action{
			LOGIN: {Name: "login appid appmm", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPID), m.Option(APPID))
				m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPMM), m.Option(APPMM))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPID)))
		}},

		"/login/": {Name: "/login/", Help: "认证", Action: map[string]*ice.Action{
			aaa.SESS: {Name: "sess code", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, WEIXIN, web.SPIDE_GET, "/sns/jscode2session?grant_type=authorization_code", "js_code", m.Option("code"),
					APPID, m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPID)), "secret", m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPMM)))

				// 用户登录
				m.Option(ice.MSG_USERZONE, MP)
				m.Echo(aaa.SessCreate(msg, msg.Append(OPENID), aaa.UserRole(msg, msg.Append(OPENID))))
			}},
			aaa.USER: {Name: "user", Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
				m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERZONE, MP, aaa.USERNICK, m.Option("nickName"),
					aaa.AVATAR, m.Option("avatarUrl"), aaa.GENDER, kit.Select("女", "男", m.Option(aaa.GENDER) == "1"),
					aaa.COUNTRY, m.Option(aaa.COUNTRY), aaa.LANGUAGE, m.Option(aaa.LANGUAGE),
					aaa.CITY, m.Option(aaa.CITY), aaa.PROVINCE, m.Option(aaa.PROVINCE),
				)
			}},
			chat.SCAN: {Name: "scan", Help: "扫码", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(chat.SCAN)
			}},
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
