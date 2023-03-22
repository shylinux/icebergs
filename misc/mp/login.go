package mp

import (
	"encoding/base64"
	"net/http"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
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
	Index.MergeCommands(ice.Commands{
		web.PP(LOGIN): {Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(web.SPIDE, mdb.CREATE, MP, mdb.Config(m, tcp.SERVER)) }},
			aaa.SESS: {Name: "sess code", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, MP, http.MethodGet, "/sns/jscode2session?grant_type=authorization_code", "js_code", m.Option(cli.CODE), APPID, mdb.Config(m, APPID), "secret", mdb.Config(m, APPMM))
				m.Option(ice.MSG_USERZONE, MP)
				m.Echo(aaa.SessCreate(msg, msg.Append(OPENID)))
			}},
			aaa.USER: {Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(aaa.USER, m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))).Length() == 0 {
					m.Cmd(aaa.USER, mdb.CREATE, m.OptionSimple(aaa.USERNAME))
				}
				m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERNICK, m.Option("nickName"), aaa.USERZONE, MP,
					aaa.AVATAR, m.Option("avatarUrl"), aaa.GENDER, kit.Select("女", "男", m.Option(aaa.GENDER) == "1"),
					m.OptionSimple(aaa.CITY, aaa.COUNTRY, aaa.LANGUAGE, aaa.PROVINCE),
				)
			}},
		}, ctx.ConfAction(tcp.SERVER, "https://api.weixin.qq.com"), aaa.WhiteAction(aaa.SESS, aaa.USER))},
		LOGIN: {Name: "login appid auto qrcode tokens create", Help: "认证", Actions: ice.Actions{
			mdb.CREATE: {Name: "create appid appmm", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				ctx.ConfigFromOption(m, APPID, APPMM)
			}},
			TOKENS: {Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); mdb.Config(m, TOKENS) == "" || now > kit.Int64(mdb.Config(m, EXPIRES)) {
					msg := m.Cmd(web.SPIDE, MP, http.MethodGet, "/cgi-bin/token?grant_type=client_credential", APPID, mdb.Config(m, APPID), "secret", mdb.Config(m, APPMM))
					if m.Warn(msg.Append(ERRCODE) != "", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
						return
					}
					mdb.Config(m, EXPIRES, now+kit.Int64(msg.Append("expires_in")))
					mdb.Config(m, TOKENS, msg.Append("access_token"))
				}
				m.Echo(mdb.Config(m, TOKENS)).Status(EXPIRES, time.Unix(kit.Int64(mdb.Config(m, EXPIRES)), 0).Format(ice.MOD_TIME))
			}},
			QRCODE: {Name: "qrcode path scene", Help: "扫码", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, MP, http.MethodPost, "/wxa/getwxacodeunlimit?access_token="+m.Cmdx(LOGIN, TOKENS), m.OptionSimple("path,scene"))
				m.Echo(kit.Format(`<img src="data:image/png;base64,%s" title='%s'>`, base64.StdEncoding.EncodeToString([]byte(msg.Result())), "some")).ProcessInner()
			}},
		}, Hand: func(m *ice.Message, arg ...string) { m.Echo(mdb.Config(m, APPID)) }},
	})
}
