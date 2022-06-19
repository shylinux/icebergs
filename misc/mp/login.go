package mp

import (
	"encoding/base64"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const (
	APPID   = "appid"
	APPMM   = "appmm"
	ACCESS  = "access"
	OPENID  = "openid"
	TOKENS  = "tokens"
	EXPIRES = "expires"
	QRCODE  = "qrcode"
)
const (
	ERRCODE = "errcode"
	ERRMSG  = "errmsg"
)
const LOGIN = "login"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		LOGIN: {Name: LOGIN, Help: "认证", Value: kit.Data(
			tcp.SERVER, "https://api.weixin.qq.com",
		)},
	}, Commands: map[string]*ice.Command{
		"/login/": {Name: "/login/", Help: "认证", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, MP, m.Config(tcp.SERVER))
			}},
			aaa.SESS: {Name: "sess code", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, MP, web.SPIDE_GET, "/sns/jscode2session?grant_type=authorization_code",
					"js_code", m.Option(cli.CODE), APPID, m.Config(APPID), "secret", m.Config(APPMM))

				// 用户登录
				m.Option(ice.MSG_USERZONE, MP)
				m.Echo(aaa.SessCreate(msg, msg.Append(OPENID)))
			}},
			aaa.USER: {Name: "user", Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
				m.Cmd(aaa.USER, mdb.MODIFY,
					aaa.USERNICK, m.Option("nickName"), aaa.USERZONE, MP,
					aaa.AVATAR, m.Option("avatarUrl"), aaa.GENDER, kit.Select("女", "男", m.Option(aaa.GENDER) == "1"),
					m.OptionSimple(aaa.CITY, aaa.COUNTRY, aaa.LANGUAGE, aaa.PROVINCE),
				)
			}},
			chat.SCAN: {Name: "scan", Help: "扫码", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(chat.GRANT) != "" {
					m.Cmdy(chat.HEADER, chat.GRANT, web.SPACE, m.Option(chat.GRANT))
					return
				}
				m.Cmdy(chat.SCAN, arg)
			}},
		}},
		LOGIN: {Name: "login appid auto qrcode tokens create", Help: "认证", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create appid appmm", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Config(APPID, m.Option(APPID))
				m.Config(APPMM, m.Option(APPMM))
			}},
			TOKENS: {Name: "tokens", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Config(TOKENS) == "" || now > kit.Int64(m.Config(EXPIRES)) {
					msg := m.Cmd(web.SPIDE, MP, web.SPIDE_GET, "/cgi-bin/token?grant_type=client_credential",
						APPID, m.Config(APPID), "secret", m.Config(APPMM))
					if m.Warn(msg.Append(ERRCODE) != "", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
						return
					}

					m.Config(EXPIRES, now+kit.Int64(msg.Append("expires_in")))
					m.Config(TOKENS, msg.Append("access_token"))
				}
				m.Echo(m.Config(TOKENS))
			}},
			QRCODE: {Name: "qrcode path scene", Help: "扫码", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, MP, web.SPIDE_POST, "/wxa/getwxacodeunlimit?access_token="+m.Cmdx(LOGIN, TOKENS),
					m.OptionSimple("path,scene"))
				m.Echo(kit.Format(`<img src="data:image/png;base64,%s" title='%s'>`, base64.StdEncoding.EncodeToString([]byte(msg.Result())), "some"))
				m.ProcessInner()
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Echo(m.Config(APPID))
		}},
	}})
}
