package wx

import (
	"crypto/sha1"
	"net/http"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat/oauth"
	kit "shylinux.com/x/toolkits"
)

const (
	CGI_BIN                     = "https://api.weixin.qq.com/cgi-bin/"
	TOKEN_CREDENTIAL            = "token?grant_type=client_credential"
	TICKET_GETTICKET            = "ticket/getticket?type=jsapi"
	QRCODE_CREATE               = "qrcode/create"
	WXACODE_UNLIMIT             = "/wxa/getwxacodeunlimit"
	MENU_CREATE                 = "menu/create"
	USER_REMARK                 = "user/info/updateremark"
	USER_INFO                   = "user/info"
	USER_GET                    = "user/get"
	USER_TAG_GET                = "user/tag/get"
	TAGS_CREATE                 = "tags/create"
	TAGS_DELETE                 = "tags/delete"
	TAGS_UPDATE                 = "tags/update"
	TAGS_GET                    = "tags/get"
	TAGS_MEMBERS_BATCHTAGGING   = "tags/members/batchtagging"
	TAGS_MEMBERS_BATCHUNTAGGING = "tags/members/batchuntagging"
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
const ACCESS = "access"

func init() {
	Index.MergeCommands(ice.Commands{
		ACCESS: {Help: "认证", Meta: Meta(), Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, WX, mdb.Config(m, tcp.SERVER))
			}},
			mdb.CREATE: {Name: "create type=web,app usernick access* appid* secret* token* icons qrcode", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(mdb.TYPE, aaa.USERNICK, ACCESS, APPID, SECRET, TOKEN, mdb.ICONS, cli.QRCODE))
				ctx.ConfigFromOption(m, ACCESS, APPID, TOKEN)
			}},
			aaa.CHECK: {Hand: func(m *ice.Message, arg ...string) {
				check := kit.Sort([]string{mdb.Config(m, TOKEN), m.Option(TIMESTAMP), m.Option(NONCE)})
				if sig := kit.Format(sha1.Sum([]byte(strings.Join(check, "")))); !m.Warn(sig != m.Option(SIGNATURE), ice.ErrNotRight, check) {
					m.Echo(ice.TRUE)
				}
			}},
			AGENT: {Hand: func(m *ice.Message, arg ...string) {
				ctx.OptionFromConfig(m, ACCESS, APPID)
			}},
			TOKENS: {Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
				if msg.Append(TOKENS) == "" || m.Time() > msg.Append(EXPIRES) {
					res := m.Cmd(web.SPIDE, WX, http.MethodGet, TOKEN_CREDENTIAL, msg.AppendSimple(APPID, SECRET))
					mdb.HashModify(m, m.OptionSimple(ACCESS), EXPIRES, m.Time(kit.Format("%vs", res.Append(oauth.EXPIRES_IN))), TOKENS, res.Append(oauth.ACCESS_TOKEN))
					msg = mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
				}
				m.Echo(msg.Append(TOKENS)).Status(msg.AppendSimple(EXPIRES))
			}},
			TICKET: {Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
				if msg.Append(TICKET) == "" || m.Time() > msg.Append(EXPIRE) {
					res := m.Cmd(web.SPIDE, WX, http.MethodGet, TICKET_GETTICKET, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS))
					mdb.HashModify(m, m.OptionSimple(ACCESS), EXPIRE, m.Time(kit.Format("%vs", res.Append(oauth.EXPIRES_IN))), TICKET, res.Append(TICKET))
					msg = mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
				}
				m.Echo(msg.Append(TICKET)).Status(msg.AppendSimple(EXPIRE))
			}},
			web.SSO: {Name: "sso name*=微信扫码 wifi env=release,trial,develop", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.CHAT_HEADER, mdb.CREATE, mdb.TYPE, mdb.PLUGIN, m.OptionSimple(mdb.NAME),
					ctx.INDEX, m.PrefixKey(), ctx.ARGS, kit.Join(kit.Simple(aaa.LOGIN, m.Option(ACCESS), m.Option(tcp.WIFI), m.Option(ENV))))
			}},
			aaa.LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd("", m.Option(ACCESS, arg[0])).Append(mdb.TYPE) == ice.WEB {
					m.Cmdy(SCAN, mdb.CREATE, mdb.TYPE, QR_STR_SCENE, mdb.NAME, m.Option(web.SPACE), mdb.TEXT, m.Option(web.SPACE),
						ctx.INDEX, web.CHAT_GRANT, ctx.ARGS, m.Option(web.SPACE))
				} else {
					h := m.Cmdx(IDE, mdb.CREATE, mdb.NAME, m.Option(web.SPACE), PAGES, PAGES_ACTION, tcp.WIFI, kit.Select("", arg, 1),
						ctx.INDEX, web.CHAT_GRANT, ctx.ARGS, kit.JoinQuery(m.OptionSimple(web.SPACE, log.DEBUG)...),
					)
					m.Echo(m.Cmdx(SCAN, UNLIMIT, SCENE, h, ENV, kit.Select("release", arg, 2), IS_HYALINE, ice.FALSE, mdb.NAME, m.Option(web.SPACE)))
				}
			}},
			web.SPACE_GRANT: {Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(ice.MSG_REFERER), "https://servicewechat.com/") {
					m.Cmd(mdb.PRUNES, m.Prefix(SCAN), "", mdb.HASH, mdb.NAME, m.Option(web.SPACE))
					m.Cmd(mdb.PRUNES, m.Prefix(IDE), "", mdb.HASH, mdb.NAME, m.Option(web.SPACE))
				}
			}},
			web.SPACE_LOGIN_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.PRUNES, m.Prefix(SCAN), "", mdb.HASH, m.OptionSimple(mdb.NAME))
				m.Cmd(mdb.PRUNES, m.Prefix(IDE), "", mdb.HASH, m.OptionSimple(mdb.NAME))
			}},
		}, aaa.RoleAction(aaa.LOGIN), gdb.EventsAction(web.SPACE_GRANT, web.SPACE_LOGIN_CLOSE), mdb.ImportantHashAction(mdb.SHORT, ACCESS, mdb.FIELD, "time,type,access,icons,usernick,appid", tcp.SERVER, CGI_BIN)), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(web.SSO, mdb.REMOVE).StatusTimeCount(mdb.ConfigSimple(m, ACCESS, APPID), web.SERVE, web.MergeLink(m, "/chat/wx/login/"))
			m.RewriteAppend(func(value, key string, index int) string {
				kit.If(key == cli.QRCODE, func() { value = ice.Render(m, ice.RENDER_QRCODE, value) })
				return value
			})
		}},
	})
}
func spidePost(m *ice.Message, api string, arg ...ice.Any) *ice.Message {
	return m.Cmd(web.SPIDE, WX, web.SPIDE_RAW, http.MethodPost, kit.MergeURL(api, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS)), arg)
}
func SpidePost(m *ice.Message, api string, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(web.SPIDE, WX, web.SPIDE_RAW, http.MethodPost, kit.MergeURL(api, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS)), arg))
}
func SpideGet(m *ice.Message, api string, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(web.SPIDE, WX, web.SPIDE_RAW, http.MethodGet, kit.MergeURL(api, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS)), arg))
}
func Meta() ice.Map {
	return kit.Dict(ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
		ACCESS, "账号", APPID, "应用", SECRET, "密码",
		TOKEN, "口令", TOKENS, "令牌", TICKET, "票据",
		EXPIRES, "令牌有效期", EXPIRE, "票据有效期", EXPIRE_SECONDS, "有效期",
		SCENE, "场景", RIVER, "一级", STORM, "二级",
		SEX, "性别", TAGS, "标签", REMARK, "备注",
		"subscribe", "订阅", "subscribe_time", "时间",
		"nickname", "昵称", "headimgurl", "头像",
		ENV, "环境", PAGES, "页面",
	)))
}
