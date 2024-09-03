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
	"shylinux.com/x/icebergs/base/nfs"
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
	WXACODE_UNLIMIT             = "/wxa/getwxacodeunlimit"
	QRCODE_CREATE               = "qrcode/create"
	MENU_CREATE                 = "menu/create"
	MEDIA_GET                   = "media/get"
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
	EXPIRE  = "expire"
	TICKET  = "ticket"
	OAUTH   = "oauth"
	MEDIA   = "media"

	STABLE_TOKEN        = "stable_token"
	STABLE_TOKEN_EXPIRE = "stable_token_expire"
)
const ACCESS = "access"

func init() {
	Index.MergeCommands(ice.Commands{
		ACCESS: {Help: "认证", Meta: Meta(), Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, mdb.Config(m, tcp.SERVER), WX, "usr/icons/wechat.png")
				m.Cmd("").Table(func(value ice.Maps) { ctx.ConfigFromOption(m.Spawn(value), ACCESS, APPID, TOKEN) })
			}},
			mdb.CREATE: {Name: "create type=web,app usernick access* appid* secret* token* icons qrcode", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(mdb.TYPE, aaa.USERNICK, ACCESS, APPID, SECRET, TOKEN, mdb.ICONS, cli.QRCODE))
				ctx.ConfigFromOption(m, ACCESS, APPID, TOKEN)
			}},
			aaa.CHECK: {Hand: func(m *ice.Message, arg ...string) {
				check := kit.Sort([]string{mdb.Config(m, TOKEN), m.Option(TIMESTAMP), m.Option(NONCE)})
				if sig := kit.Format(sha1.Sum([]byte(strings.Join(check, "")))); !m.WarnNotRight(sig != m.Option(SIGNATURE), check) {
					m.Echo(ice.TRUE)
				}
			}},
			STABLE_TOKEN: {Hand: func(m *ice.Message, arg ...string) {
				spideToken(m, STABLE_TOKEN, STABLE_TOKEN, STABLE_TOKEN_EXPIRE, "grant_type", "client_credential")
			}},
			TOKENS: {Hand: func(m *ice.Message, arg ...string) {
				spideToken(m, TOKEN_CREDENTIAL, TOKENS, EXPIRES)
			}},
			TICKET: {Hand: func(m *ice.Message, arg ...string) {
				spideToken(m, TICKET_GETTICKET, TICKET, EXPIRE, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS))
			}},
			AGENT: {Hand: func(m *ice.Message, arg ...string) {
				ctx.OptionFromConfig(m, ACCESS, APPID)
			}},
			"user": {Name: "user openid", Hand: func(m *ice.Message, arg ...string) {
				SpideGet(m, "user/info", OPENID, m.Option(OPENID))
			}},
			"api": {Name: "api method=GET,POST path params", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option("method") {
				case "POST":
					res := SpidePost(m, m.Option(nfs.PATH), kit.Split(m.Option("params")))
					m.Echo(kit.Formats(res))
				case "GET":
					res := SpideGet(m, m.Option(nfs.PATH), kit.Split(m.Option("params")))
					m.Echo(kit.Formats(res))
				}
			}},
			MEDIA: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPIDE, WX, web.SPIDE_SAVE, arg[1], http.MethodGet, MEDIA_GET, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS), "media_id", arg[0])
			}},
			OAUTH: {Icon: "bi bi-shield-fill-check", Hand: func(m *ice.Message, arg ...string) {
				oauth := m.Cmdx("web.chat.oauth.client", mdb.CREATE,
					"domain", "https://api.weixin.qq.com", "client_id", m.Option(APPID), "client_secret", m.Option(SECRET),
					"oauth_url", "https://open.weixin.qq.com/connect/oauth2/authorize?appid="+m.Option(APPID), "scope", "snsapi_userinfo", "login", "login2",
					"grant_url", "/sns/oauth2/access_token",
					"token_url", "/sns/oauth2/refresh_token",
					"users_url", "/sns/userinfo", "user_key", "openid", "nick_key", "nickname", "icon_key", "headimgurl",
					m.OptionSimple("user_cmd", "sess_cmd"),
				)
				m.Cmd(AGENT, OAUTH, m.Cmdx("web.chat.oauth.client", web.LINK, oauth))
				m.Cmd(web.SPACE, ice.OPS, ctx.CONFIG, "web.chat.header", OAUTH, m.Cmdx("web.chat.oauth.client", web.LINK, oauth))
			}},
			web.SSO: {Name: "sso name*=weixin help*=微信扫码 order=11 env=release,trial,develop wifi", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.CHAT_HEADER, mdb.CREATE, mdb.TYPE, mdb.PLUGIN, m.OptionSimple(mdb.NAME, mdb.HELP, mdb.ORDER),
					ctx.INDEX, m.PrefixKey(), ctx.ARGS, kit.Join(kit.Simple(aaa.LOGIN, m.Option(ACCESS), m.Option(ENV), m.Option(tcp.WIFI))))
			}},
			aaa.LOGIN: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd("", m.Option(ACCESS, arg[0])).Append(mdb.TYPE) == ice.WEB {
					m.Cmdy(SCAN, mdb.CREATE, mdb.TYPE, QR_STR_SCENE, mdb.NAME, "请授权登录", mdb.TEXT, m.Option(web.SPACE),
						ctx.INDEX, web.CHAT_GRANT, ctx.ARGS, m.Option(web.SPACE))
				} else {
					h := m.Cmdx(IDE, mdb.CREATE, mdb.NAME, "请授权登录", mdb.TEXT, m.Option(web.SPACE), PAGES, PAGES_ACTION, tcp.WIFI, kit.Select("", arg, 2),
						ctx.INDEX, web.CHAT_GRANT, ctx.ARGS, kit.JoinQuery(m.OptionSimple(web.SPACE, log.DEBUG)...))
					m.Echo(m.Cmdx(SCAN, UNLIMIT, ENV, kit.Select("develop", arg, 1), SCENE, h, mdb.NAME, m.Option(web.SPACE)))
				}
			}},
			web.SPACE_GRANT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.PRUNES, m.Prefix(SCAN), "", mdb.HASH, mdb.TEXT, m.Option(web.SPACE))
				m.Cmd(mdb.PRUNES, m.Prefix(SCAN), "", mdb.HASH, mdb.NAME, m.Option(web.SPACE))
				m.Cmd(mdb.PRUNES, m.Prefix(IDE), "", mdb.HASH, mdb.TEXT, m.Option(web.SPACE))
				m.Cmd(mdb.PRUNES, web.SHARE, "", mdb.HASH, mdb.TEXT, m.Option(web.SPACE))
			}},
			web.SPACE_LOGIN_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.PRUNES, m.Prefix(SCAN), "", mdb.HASH, m.OptionSimple(mdb.NAME))
				m.Cmd(mdb.PRUNES, m.Prefix(IDE), "", mdb.HASH, m.OptionSimple(mdb.NAME))
			}},
		}, gdb.EventsAction(web.SPACE_GRANT, web.SPACE_LOGIN_CLOSE), mdb.ExportHashAction(
			mdb.SHORT, ACCESS, mdb.FIELD, "time,type,access,icons,usernick,appid,secret,token", tcp.SERVER, CGI_BIN,
		)), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(OAUTH, web.SSO, TICKET, TOKENS, STABLE_TOKEN, "user", "api", mdb.REMOVE).StatusTimeCount(mdb.ConfigSimple(m, ACCESS, APPID), web.SERVE, m.MergeLink("/chat/wx/login/"))
			m.RewriteAppend(func(value, key string, index int) string {
				kit.If(key == cli.QRCODE, func() { value = ice.Render(m, ice.RENDER_QRCODE, value) })
				return value
			})
		}},
	})
}
func spideToken(m *ice.Message, api string, token, expire string, arg ...string) {
	msg := mdb.HashSelect(m.Spawn(), m.OptionDefault(ACCESS, mdb.Config(m, ACCESS)))
	m.Info("what token %v %v", msg.Append(expire), msg.Append(token))
	if msg.Append(token) == "" || m.Time() > msg.Append(expire) {
		kit.If(api != TICKET_GETTICKET, func() { arg = append(arg, msg.AppendSimple(APPID, SECRET)...) })
		res := m.Cmd(web.SPIDE, WX, kit.Select(http.MethodGet, http.MethodPost, api == STABLE_TOKEN), api, arg)
		if m.Warn(!kit.IsIn(res.Append("errcode"), "0", ""), res.Append("errmsg")) {
			return
		}
		m.Info("what res: %v", res.FormatMeta())
		mdb.HashModify(m, m.OptionSimple(ACCESS), expire, m.Time(kit.Format("%vs", res.Append(oauth.EXPIRES_IN))), token, res.Append(kit.Select(oauth.ACCESS_TOKEN, TICKET, api == TICKET_GETTICKET)))
		msg = mdb.HashSelect(m.Spawn(), m.Option(ACCESS))
	}
	m.Echo(msg.Append(token)).Status(msg.AppendSimple(expire))
}
func spidePost(m *ice.Message, api string, arg ...ice.Any) *ice.Message {
	return m.Cmd(web.SPIDE, WX, web.SPIDE_RAW, http.MethodPost, kit.MergeURL(api, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS)), arg)
}
func SpidePost(m *ice.Message, api string, arg ...ice.Any) ice.Any {
	res := kit.UnMarshal(m.Cmdx(web.SPIDE, WX, web.SPIDE_RAW, http.MethodPost, kit.MergeURL(api, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS)), arg))
	m.Warn(!kit.IsIn(kit.Format(kit.Value(res, "errcode")), "", "0"), kit.Value(res, "errmsg"))
	m.Info("res: %v", kit.Format(res))
	return res
}
func SpideGet(m *ice.Message, api string, arg ...ice.Any) ice.Any {
	res := kit.UnMarshal(m.Cmdx(web.SPIDE, WX, web.SPIDE_RAW, http.MethodGet, kit.MergeURL(api, oauth.ACCESS_TOKEN, m.Cmdx(ACCESS, TOKENS)), arg))
	m.Warn(!kit.IsIn(kit.Format(kit.Value(res, "errcode")), "", "0"), kit.Value(res, "errmsg"))
	m.Info("res: %v", kit.Format(res))
	return res
}
func Meta() ice.Map {
	return kit.Dict(ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
		ACCESS, "账号", APPID, "应用", SECRET, "密码", TOKEN, "口令",
		TOKENS, "令牌", EXPIRES, "令牌有效期",
		TICKET, "票据", EXPIRE, "票据有效期", EXPIRE_SECONDS, "有效期",
		SCENE, "场景", RIVER, "一级", STORM, "二级",
		SEX, "性别", TAGS, "标签", REMARK, "备注",
		"subscribe", "订阅", "subscribe_time", "时间",
		"nickname", "昵称", "headimgurl", "头像",
		"projectname", "项目",
		ENV, "环境", PAGES, "页面",
	), html.VALUE, kit.Dict(
		"web", "公众号", "app", "小程序",
	)))
}
