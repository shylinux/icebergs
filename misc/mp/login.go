package mp

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const (
	APPID  = "appid"
	APPMM  = "appmm"
	ACCESS = "access"
	OPENID = "openid"
)
const LOGIN = "login"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		LOGIN: {Name: LOGIN, Help: "认证", Value: kit.Data(
			tcp.SERVER, "https://api.weixin.qq.com",
			APPID, "", APPMM, "", "tokens", "",
		)},
	}, Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(web.SPIDE, mdb.CREATE, MP, m.Conf(LOGIN, kit.Keym(tcp.SERVER)))
		}},
		"/login/": {Name: "/login/", Help: "认证", Action: map[string]*ice.Action{
			aaa.SESS: {Name: "sess code", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, MP, web.SPIDE_GET, "/sns/jscode2session?grant_type=authorization_code",
					"js_code", m.Option(kit.MDB_CODE), APPID, m.Config(APPID), "secret", m.Config(APPMM))

				// 用户登录
				m.Option(ice.MSG_USERZONE, MP)
				m.Echo(aaa.SessCreate(msg, msg.Append(OPENID)))
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
				m.Cmdy(chat.SCAN, arg)
			}},
		}},
		LOGIN: {Name: "login appid auto login", Help: "认证", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create appid appmm", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Config(APPID, m.Option(APPID))
				m.Config(APPMM, m.Option(APPMM))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(m.Config(APPID))
		}},
	},
	})
}
