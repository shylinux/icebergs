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
	APPID  = "appid"
	APPMM  = "appmm"
	ACCESS = "access"
	OPENID = "openid"
	WEIXIN = "weixin"
)
const LOGIN = "login"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			LOGIN: {Name: LOGIN, Help: "认证", Value: kit.Data(
				WEIXIN, "https://api.weixin.qq.com", APPID, "", APPMM, "",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, WEIXIN, m.Conf(LOGIN, kit.Keym(WEIXIN)))
			}},
			"/login/": {Name: "/login/", Help: "认证", Action: map[string]*ice.Action{
				aaa.SESS: {Name: "sess code", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(web.SPIDE, WEIXIN, web.SPIDE_GET, "/sns/jscode2session?grant_type=authorization_code",
						"js_code", m.Option(kit.MDB_CODE), APPID, m.Conf(LOGIN, kit.Keym(APPID)), "secret", m.Conf(LOGIN, kit.Keym(APPMM)))

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
					m.Conf(LOGIN, kit.Keym(APPID), m.Option(APPID))
					m.Conf(LOGIN, kit.Keym(APPMM), m.Option(APPMM))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(m.Conf(LOGIN, kit.Keym(APPID)))
			}},
		},
	})
}
