package wx

import (
	"crypto/sha1"
	"net/http"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat/oauth"
	kit "shylinux.com/x/toolkits"
)

const (
	APPID   = "appid"
	SECRET  = "secret"
	TOKEN   = "token"
	TOKENS  = "tokens"
	EXPIRES = "expires"
	TICKET  = "ticket"
	EXPIRE  = "expire"
)
const (
	CGI_BIN       = "https://api.weixin.qq.com/cgi-bin/"
	QRCODE_CREATE = "qrcode/create"
	MENU_CREATE   = "menu/create"
	USER_INFO     = "user/info"
	USER_GET      = "user/get"
)
const ACCESS = "access"

func init() {
	Index.MergeCommands(ice.Commands{
		ACCESS: {Help: "认证", Meta: Meta(), Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(web.SPIDE, mdb.CREATE, WX, mdb.Config(m, tcp.SERVER)) }},
			mdb.CREATE: {Name: "login usernick access* appid* secret* token* icons", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(aaa.USERNICK, ACCESS, APPID, SECRET, TOKEN, mdb.ICONS))
				ctx.ConfigFromOption(m, ACCESS, APPID, TOKEN)
			}},
			aaa.CHECK: {Hand: func(m *ice.Message, arg ...string) {
				check := kit.Sort([]string{mdb.Config(m, TOKEN), m.Option(TIMESTAMP), m.Option(NONCE)})
				if sig := kit.Format(sha1.Sum([]byte(strings.Join(check, "")))); !m.Warn(sig != m.Option(SIGNATURE), ice.ErrNotRight, check) {
					m.Echo(ice.TRUE)
				}
			}},
			AGENT: {Hand: func(m *ice.Message, arg ...string) { ctx.OptionFromConfig(m, ACCESS, APPID) }},
			TOKENS: {Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
				if msg.Append(TOKENS) == "" || m.Time() > msg.Append(EXPIRES) {
					res := m.Cmd(web.SPIDE, WX, http.MethodGet, "token?grant_type=client_credential", msg.AppendSimple(APPID, SECRET))
					mdb.HashModify(m, m.OptionSimple(ACCESS), EXPIRES, m.Time(kit.Format("%vs", res.Append(oauth.EXPIRES_IN))), TOKENS, res.Append(oauth.ACCESS_TOKEN))
					msg = mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
				}
				m.Echo(msg.Append(TOKENS)).Status(msg.AppendSimple(EXPIRES))
			}},
			TICKET: {Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
				if msg.Append(TICKET) == "" || m.Time() > msg.Append(EXPIRE) {
					res := m.Cmd(web.SPIDE, WX, http.MethodGet, "ticket/getticket?type=jsapi", arg, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS))
					mdb.HashModify(m, m.OptionSimple(ACCESS), EXPIRE, m.Time(kit.Format("%vs", res.Append(oauth.EXPIRES_IN))), TICKET, res.Append(TICKET))
					msg = mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
				}
				m.Echo(msg.Append(TICKET)).Status(msg.AppendSimple(EXPIRE))
			}},
		}, mdb.ImportantHashAction(mdb.SHORT, ACCESS, mdb.FIELD, "time,access,usernick,appid,icons", tcp.SERVER, CGI_BIN)), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).StatusTimeCount(mdb.ConfigSimple(m, ACCESS, APPID), web.LINK, web.MergeURL2(m, "/chat/wx/login/"))
		}},
	})
}
func SpideGet(m *ice.Message, api string, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(web.SPIDE, WX, web.SPIDE_RAW, http.MethodGet, kit.MergeURL(api, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS)), arg))
}
func SpidePost(m *ice.Message, api string, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(web.SPIDE, WX, web.SPIDE_RAW, http.MethodPost, kit.MergeURL(api, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS)), arg))
}
func Meta() ice.Map {
	return kit.Dict(ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
		ACCESS, "账号", APPID, "应用", SECRET, "密码",
		EXPIRE_SECONDS, "有效期",
		SCENE, "场景", RIVER, "一级", STORM, "二级",
	)))
}
