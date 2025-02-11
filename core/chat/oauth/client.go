package oauth

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	OAUTH         = "oauth"
	DOMAIN        = "domain"
	CLIENT_ID     = "client_id"
	CLIENT_SECRET = "client_secret"

	OAUTH_URL = "oauth_url"
	GRANT_URL = "grant_url"
	TOKEN_URL = "token_url"
	USERS_URL = "users_url"
	USER_KEY  = "user_key"
	NICK_KEY  = "nick_key"
	ICON_KEY  = "icon_key"
	SESS_CMD  = "sess_cmd"
	USER_CMD  = "user_cmd"

	REDIRECT_URI       = "redirect_uri"
	RESPONSE_TYPE      = "response_type"
	AUTHORIZATION_CODE = "authorization_code"
	GRANT_TYPE         = "grant_type"
	STATE              = "state"
	SCOPE              = "scope"
	CODE               = "code"

	API_PREFIX   = "api_prefix"
	TOKEN_PREFIX = "token_prefix"
	ACCESS_TOKEN = "access_token"
	EXPIRES_IN   = "expires_in"
)

type Client struct {
	ice.Hash
	short  string `data:"domain,client_id"`
	field  string `data:"time,hash,domain,client_id,client_secret,oauth_url,grant_url,token_url,users_url,scope,login,user_key,user_cmd,sess_cmd,nick_key,icon_key,api_prefix,token_prefix"`
	sso    string `name:"sso name* help icons*" help:"登录"`
	auth   string `name:"auth" help:"授权" icon:"bi bi-person-check"`
	user   string `name:"user" help:"用户" icon:"bi bi-person-vcard"`
	orgs   string `name:"orgs" help:"组织"`
	repo   string `name:"repo" help:"资源"`
	list   string `name:"list hash auto" help:"授权" icon:"oauth.png"`
	login  string `name:"login" role:"void"`
	login2 string `name:"login2" role:"void"`
}

var Inputs = map[string]map[string]string{}

func init() {
	Inputs["repos"] = map[string]string{
		OAUTH_URL:    "/login/oauth/authorize",
		GRANT_URL:    "/login/oauth/access_token",
		TOKEN_URL:    "/login/oauth/access_token",
		USERS_URL:    "/api/v1/user",
		API_PREFIX:   "/api/v1/",
		TOKEN_PREFIX: "token",
	}
}
func (s Client) Config(m *ice.Message, arg ...string) {
	s.Create(m, kit.Simple(Inputs[arg[1]], arg, web.DOMAIN, m.Cmdv(web.SPIDE, arg[1], web.CLIENT_ORIGIN))...)
}
func (s Client) Inputs(m *ice.Message, arg ...string) {
	switch m.Option(ctx.ACTION) {
	case web.SSO:
		switch arg[0] {
		case mdb.NAME:
		case mdb.ICON:
		}
	}
	switch s.Hash.Inputs(m, arg...); arg[0] {
	case web.DOMAIN:
		m.Cmdy(web.SPIDE, mdb.INPUTS, arg)
	default:
		for _, input := range Inputs {
			if value, ok := input[arg[0]]; ok {
				m.Push(arg[0], value)
			}
		}
	}
}
func (s Client) List(m *ice.Message, arg ...string) {
	s.Hash.List(m, arg...).PushAction(s.User, s.Auth, s.Sso, s.Remove).EchoScript(s.RedirectURI(m))
}
func (s Client) Sso(m *ice.Message, arg ...string) {
	m.Cmd(web.CHAT_HEADER, mdb.CREATE, OAUTH, m.Option(mdb.NAME), m.Option(mdb.HELP), m.Option(mdb.ICONS), s.OAuthURL(m))
}
func (s Client) Auth(m *ice.Message, arg ...string) {
	m.ProcessOpen(s.OAuthURL(m))
}
func (s Client) Link(m *ice.Message, arg ...string) {
	m.Options(m.Cmd("", arg[0]).AppendSimple())
	m.Echo(s.OAuthURL(m))
}
func (s Client) User(m *ice.Message, arg ...string) {
	if res := s.Get(m, m.Option(mdb.HASH), m.Option(USERS_URL), arg...); res != nil {
		if m.Options(res); m.Warn(!kit.IsIn(m.Option("errcode"), "", "0"), m.Option("errmsg")) {
			return
		}
		m.Info("user info %v", kit.Format(res))
		if m.Option(USER_CMD) != "" {
			m.Options("open_id", m.Option("openid"), aaa.USERNICK, m.Option("nickname"), aaa.AVATAR, m.Option("headimgurl"))
			m.Cmdy(kit.Split(m.Option(USER_CMD)), m.OptionSimple("open_id"), kit.Simple(res))
			return
		}
		username := m.Option(aaa.USERNAME, m.Option(kit.Select(aaa.USERNAME, m.Option(USER_KEY))))
		if m.Cmd(aaa.USER, username).Length() > 0 {
			m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERNAME, username,
				aaa.USERNICK, m.Option(kit.Select("full_name", m.Option(NICK_KEY))),
				aaa.AVATAR, m.Option(kit.Select(aaa.AVATAR_URL, m.Option(ICON_KEY))),
			)
		} else {
			m.Cmd(aaa.USER, mdb.CREATE,
				aaa.USERROLE, kit.Select(aaa.VOID, aaa.TECH, m.Option("is_admin") == ice.TRUE),
				aaa.USERNAME, username,
				aaa.USERNICK, m.Option(kit.Select("full_name", m.Option(NICK_KEY))),
				aaa.USERZONE, m.Option(web.DOMAIN),
				aaa.AVATAR, m.Option(kit.Select(aaa.AVATAR_URL, m.Option(ICON_KEY))),
				m.OptionSimple(aaa.LANGUAGE, aaa.EMAIL))
		}
	}
}
func (s Client) Orgs(m *ice.Message, arg ...string) {}
func (s Client) Repo(m *ice.Message, arg ...string) {}

func init() { ice.Cmd("web.chat.oauth.client", Client{}) }

func (s Client) Login(m *ice.Message, arg ...string) {
	if state, code := m.Option(STATE), m.Option(CODE); !m.WarnNotValid(state == "" || code == "") {
		s.Hash.List(m.Spawn(), m.Option(mdb.HASH, state)).Table(func(value ice.Maps) { m.Options(value) })
		m.Options(GRANT_TYPE, AUTHORIZATION_CODE, REDIRECT_URI, s.RedirectURI(m))
		if res := s.Post(m, m.Option(mdb.HASH), m.Option(GRANT_URL), m.OptionSimple(GRANT_TYPE, CODE, CLIENT_ID, CLIENT_SECRET, REDIRECT_URI)...); !m.WarnNotValid(res == nil) {
			kit.Value(res, EXPIRES_IN, m.Time(kit.Format("%vs", kit.Int(kit.Value(res, EXPIRES_IN)))))
			m.Options(res)
			if s.User(m); !m.WarnNotValid(m.Option(aaa.USERNAME) == "") && m.R != nil {
				m.Cmd(aaa.USER, mdb.MODIFY, m.OptionSimple(aaa.USERNAME), kit.Simple(res))
				web.RenderCookie(m.Message, aaa.SessCreate(m.Message, m.Option(aaa.USERNAME)))
				m.ProcessHistory()
			} else {
				m.ProcessClose()
			}
		}
	}
}
func (s Client) Login2(m *ice.Message, arg ...string) {
	if state, code := m.Option(STATE), m.Option(CODE); !m.WarnNotValid(state == "" || code == "") {
		msg := m.Spawn()
		s.Hash.List(msg, m.Option(mdb.HASH, state)).Table(func(value ice.Maps) { msg.Options(value) })
		msg.Options(GRANT_TYPE, AUTHORIZATION_CODE, REDIRECT_URI, s.RedirectURI(msg)).Option(ACCESS_TOKEN, "")
		if res := s.Get(msg, msg.Option(mdb.HASH), msg.Option(GRANT_URL), kit.Simple(msg.OptionSimple(GRANT_TYPE, CODE), "appid", msg.Option(CLIENT_ID), "secret", msg.Option(CLIENT_SECRET))...); !m.WarnNotValid(res == nil) {
			if msg.Options(res); m.Warn(!kit.IsIn(msg.Option("errcode"), "", "0"), msg.Option("errmsg")) {
				return
			}
			m.Info("token info %v", kit.Format(res))
			msg.Option(EXPIRES_IN, m.Time(kit.Format("%vs", kit.Int(msg.Option(EXPIRES_IN)))))
			if s.User(msg, msg.OptionSimple("openid")...); !m.Warn(msg.Option(aaa.USERNAME) == "" && msg.Option("user_uid") == "") {
				if msg.Option(SESS_CMD) != "" {
					m.Cmdy(kit.Split(msg.Option(SESS_CMD)), kit.Dict("user_uid", msg.Option("user_uid")))
				} else {
					m.ProcessCookie(ice.MSG_SESSID, aaa.SessCreate(m.Message, msg.Option(aaa.USERNAME)), "-2")
				}
			}
		}
	}
}
func (s Client) OAuthURL(m *ice.Message) string {
	return kit.MergeURL2(m.Option(web.DOMAIN), m.Option(OAUTH_URL), RESPONSE_TYPE, CODE, m.OptionSimple(CLIENT_ID), REDIRECT_URI, s.RedirectURI(m), m.OptionSimple(SCOPE), STATE, m.Option(mdb.HASH))
}
func (s Client) RedirectURI(m *ice.Message) string {
	return m.MergeLink(web.ChatCmdPath(m.Message, m.ShortKey(), ctx.ACTION, kit.Select(aaa.LOGIN, m.Option(aaa.LOGIN))), log.DEBUG, m.Option(log.DEBUG))
}

func (s Client) Get(m *ice.Message, hash, api string, arg ...string) ice.Any {
	return web.SpideGet(m.Message, s.request(m, hash, api, arg...))
}
func (s Client) Put(m *ice.Message, hash, api string, arg ...string) ice.Any {
	return web.SpidePut(m.Message, s.request(m, hash, api, arg...))
}
func (s Client) Post(m *ice.Message, hash, api string, arg ...string) ice.Any {
	return web.SpidePost(m.Message, s.request(m, hash, api, arg...))
}
func (s Client) Delete(m *ice.Message, hash, api string, arg ...string) ice.Any {
	return web.SpideDelete(m.Message, s.request(m, hash, api, arg...))
}
func (s Client) Save(m *ice.Message, hash, file, api string, arg ...string) ice.Any {
	args := s.request(m, hash, api, arg...)
	web.Toast(m.Message, "process", file, "-1", 0)
	return web.SpideSave(m.Message, file, kit.MergeURL(args[0], args[1:]), func(count int, total int, value int) {
		web.Toast(m.Message, kit.Format("%s/%s", kit.FmtSize(count), kit.FmtSize(total)), file, "-1", value)
	})
}
func (s Client) request(m *ice.Message, hash, api string, arg ...string) []string {
	msg := s.Hash.List(m.Spawn(), hash)
	kit.If(m.Option(ACCESS_TOKEN) == "" && m.Option(ice.MSG_USERNAME) != "", func() {
		msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
		m.Option(ACCESS_TOKEN, msg.Append(ACCESS_TOKEN))
	})
	kit.If(m.Option(ACCESS_TOKEN), func(p string) {
		m.Options(web.SPIDE_HEADER, ice.Maps{html.Authorization: msg.Append(TOKEN_PREFIX) + lex.SP + p})
		arg = append(arg, ACCESS_TOKEN, p)
	})
	kit.If(api == "", func() { api = path.Join(msg.Append(API_PREFIX), m.ActionKey()) })
	return kit.Simple(kit.MergeURL2(msg.Append(web.DOMAIN), api), arg)
}

func ClientCreate(m *ice.Message, domain, client_id, client_secret string, arg ...string) {
	m.AdminCmd("web.chat.oauth.client", mdb.CREATE, DOMAIN, domain, CLIENT_ID, client_id, CLIENT_SECRET, client_secret, arg)
}
