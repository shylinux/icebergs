package chat

import (
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

func _merge_url(m *ice.Message, domain, key string, arg ...ice.Any) string {
	if domain == "" {
		if m.Option(ice.MSG_USERPOD) == "" {
			domain = m.MergeLink(ice.PS)
		} else {
			domain = m.MergeLink("/chat/pod/" + m.Option(ice.MSG_USERPOD))
		}
	}
	if domain = strings.TrimSuffix(domain, ice.PS); strings.Contains(domain, "/chat/pod/") {
		domain += web.P(strings.TrimPrefix(m.Prefix(web.P(key)), "web.chat."))
	} else {
		domain += m.RoutePath(key)
	}
	return kit.MergeURL(domain, arg...)
}

const (
	CHECK = "check"
	APPLY = "apply"
	REPLY = "reply"
	OFFER = "offer"
)
const (
	ACCESS_TOKEN = "access_token"
	TOKEN_TYPE   = "token_type"
	EXPIRES_IN   = "expires_in"
	CODE         = "code"
	STATE        = "state"
	SCOPE        = "scope"
	CLIENT_ID    = "client_id"
	REDIRECT_URI = "redirect_uri"
)
const (
	AUTHORIZE = "authorize"
	TOKEN     = "token"
	ACCESS    = "access"
	USERINFO  = "userinfo"
)
const OAUTH = "oauth"

var Index = &ice.Context{Name: OAUTH, Help: "认证授权", Commands: map[string]*ice.Command{
	OAUTH: {Name: "oauth hash auto prunes", Help: "权限", Action: ice.MergeAction(map[string]*ice.Action{
		CHECK: {Name: "check scope domain", Help: "检查", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(_merge_url(m, kit.Select(ice.Info.Make.Domain, m.Option(web.DOMAIN)), APPLY, m.OptionSimple(SCOPE), REDIRECT_URI, _merge_url(m, "", REPLY)))
		}},
		APPLY: {Name: "apply scope redirect_uri", Help: "申请", Hand: func(m *ice.Message, arg ...string) {
			if m.Right(m.Option(SCOPE)) {
				token := m.Cmdx(OFFER, mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.OptionSimple(SCOPE, REDIRECT_URI))
				m.ProcessReplace(m.Option(REDIRECT_URI), m.OptionSimple(SCOPE), OFFER, _merge_url(m, "", OFFER, ACCESS_TOKEN, token))
			} else {
				m.Cmdy(APPLY, mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.OptionSimple(SCOPE, REDIRECT_URI))
			}
		}},
		REPLY: {Name: "reply scope offer", Help: "通过", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(REPLY, mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.OptionSimple(SCOPE, OFFER))

			m.Option(web.SPIDE_HEADER, web.UserAgent, m.PrefixKey())
			m.Cmd(ssh.SOURCE, m.Option(SCOPE), kit.Dict(nfs.CAT_CONTENT, m.Cmdx(web.SPIDE, ice.DEV, web.SPIDE_GET, m.Option(OFFER))))
			m.ProcessHistory()
		}},
	})},

	APPLY: {Name: "apply hash auto create prunes", Help: "申请", Action: mdb.HashAction(mdb.EXPIRE, "72h", mdb.FIELD, "time,hash,username,scope,redirect_uri")},
	REPLY: {Name: "reply hash auto create prunes", Help: "授权", Action: mdb.HashAction(mdb.EXPIRE, "720h", mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,username,scope,offer")},
	OFFER: {Name: "offer hash auto create prunes", Help: "访问", Action: mdb.HashAction(mdb.EXPIRE, "720h", mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,username,scope,redirect_uri")},

	web.P(APPLY): {Name: "/apply scope redirect_uri", Help: "申请", Action: ctx.CmdAction(), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if m.Option(REDIRECT_URI) == "" {
			m.RenderStatusBadRequest() // 参数错误

		} else { // 申请
			m.RenderCmd(m.Prefix(OAUTH), APPLY)
		}
	}},
	web.P(REPLY): {Name: "/reply scope offer", Help: "授权", Action: ctx.CmdAction(), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if m.Option(OFFER) == "" {
			m.RenderStatusBadRequest() // 参数错误

		} else { // 授权
			m.RenderCmd(m.Prefix(OAUTH), REPLY)
		}
	}},
	web.P(OFFER): {Name: "/offer access_token", Help: "访问", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if m.Option(ACCESS_TOKEN) == "" {
			m.RenderStatusBadRequest() // 参数错误

		} else if msg := m.Cmd(OFFER, m.Option(ACCESS_TOKEN), ice.OptionFields("time,scope")); kit.Time(msg.Append(mdb.TIME)) < kit.Time(msg.Time()) {
			m.RenderStatusUnauthorized() // 已过期

		} else { // 访问
			aaa.UserRoot(m).Cmdy(nfs.CAT, msg.Append(SCOPE)).RenderResult()
		}
	}},

	AUTHORIZE: {Name: "authorize hash auto create prunes", Help: "认证", Action: mdb.HashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,redirect_uri")},
	TOKEN:     {Name: "token hash auto create prunes", Help: "授权", Action: mdb.HashAction(mdb.EXPIRE, "72h", mdb.FIELD, "time,hash,used,state,scope,redirect_uri")},
	ACCESS:    {Name: "access hash auto create prunes", Help: "访问", Action: mdb.HashAction(mdb.EXPIRE, "720h", mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,username,scope,redirect_uri")},

	web.P(AUTHORIZE): {Name: "/authorize state scope client_id redirect_uri", Help: "认证", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if m.Option(CLIENT_ID) == "" || m.Option(REDIRECT_URI) == "" {
			m.RenderStatusBadRequest() // 参数错误

		} else if uri := m.Cmd(AUTHORIZE, m.Option(CLIENT_ID)).Append(REDIRECT_URI); m.Warn(uri == "", ice.ErrNotFound, CLIENT_ID) {
			m.RenderStatusNotFound() // 未找到

		} else if m.Warn(!strings.HasPrefix(m.Option(REDIRECT_URI), uri), ice.ErrNotRight, REDIRECT_URI) {
			m.RenderStatusForbidden() // 未匹配

		} else { // 认证
			m.RenderRedirect(m.Option(REDIRECT_URI), CODE, m.Cmdx(TOKEN, mdb.CREATE, m.OptionSimple(STATE, SCOPE, REDIRECT_URI)), m.OptionSimple(STATE))
		}
	}},
	web.P(TOKEN): {Name: "/token code redirect_uri", Help: "授权", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if m.Option(CODE) == "" || m.Option(REDIRECT_URI) == "" {
			m.RenderStatusBadRequest() // 参数错误
			return
		}

		const USED = "used"
		msg := m.Cmd(TOKEN, m.Option(CODE))
		if uri := msg.Append(REDIRECT_URI); m.Warn(uri == "", ice.ErrNotFound, CODE) {
			m.RenderStatusNotFound() // 未找到

		} else if m.Warn(!strings.HasPrefix(m.Option(REDIRECT_URI), uri), ice.ErrNotRight, REDIRECT_URI) {
			m.RenderStatusForbidden() // 未匹配

		} else if m.Warn(msg.Append(USED) == ice.TRUE, ice.ErrNotRight, CODE) {
			m.RenderStatusForbidden() // 已使用

		} else if kit.Time(msg.Append(mdb.TIME)) < kit.Time(m.Time()) {
			m.RenderStatusUnauthorized() // 已过期

		} else { // 授权
			token := m.Cmdx(ACCESS, mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), msg.AppendSimple(SCOPE, REDIRECT_URI))
			m.RenderJson(ACCESS_TOKEN, token, TOKEN_TYPE, web.Bearer, EXPIRES_IN, kit.Duration(m.Conf(ACCESS, kit.Keym(mdb.EXPIRE)))/time.Second)
			m.Cmdx(TOKEN, mdb.MODIFY, mdb.HASH, m.Option(CODE), USED, ice.TRUE)
		}
	}},
	web.P(USERINFO): {Name: "/userinfo Authorization", Help: "信息", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if ls := strings.SplitN(m.R.Header.Get(web.Authorization), ice.SP, 2); m.Warn(len(ls) != 2 || ls[1] == "", ice.ErrNotFound, web.Bearer) {
			m.RenderStatusBadRequest() // 参数错误

		} else if msg := m.Cmd(ACCESS, ls[1]); kit.Time(msg.Append(mdb.TIME)) < kit.Time(m.Time()) {
			m.RenderStatusUnauthorized() // 已过期

		} else { // 访问
			m.RenderJson(mdb.NAME, msg.Append(aaa.USERNAME), aaa.EMAIL, msg.Append(aaa.USERNAME))
		}
	}},
}}

func init() { chat.Index.Register(Index, &web.Frame{}) }
