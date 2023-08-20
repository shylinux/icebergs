package oauth

import (
	"path"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	CLIENT_ID     = "client_id"
	CLIENT_SECRET = "client_secret"

	OAUTH_URL = "oauth_url"
	GRANT_URL = "grant_url"
	TOKEN_URL = "token_url"
	USERS_URL = "users_url"

	REDIRECT_URI       = "redirect_uri"
	RESPONSE_TYPE      = "response_type"
	AUTHORIZATION_CODE = "authorization_code"
	GRANT_TYPE         = "grant_type"
	STATE              = "state"
	CODE               = "code"
	API                = "api"

	ACCESS_TOKEN = "access_token"
	EXPIRES_IN   = "expires_in"
)

type Client struct {
	ice.Hash
	short string `data:"domain,client_id"`
	field string `data:"time,hash,domain,client_id,client_secret,oauth_url,grant_url,token_url,users_url,api,prefix"`
	auth  string `name:"auth" help:"授权"`
	user  string `name:"user" help:"用户"`
	orgs  string `name:"orgs" help:"组织"`
	repo  string `name:"repo" help:"源码"`
	list  string `name:"list hash auto" help:"授权"`
}

func (s Client) Init(m *ice.Message, arg ...string) {
	aaa.White(m.Message, m.PrefixKey(), ctx.ACTION, aaa.LOGIN)
	s.Hash.Init(m, arg...)
}
func (s Client) Inputs(m *ice.Message, arg ...string) {
	switch s.Hash.Inputs(m, arg...); arg[0] {
	case web.DOMAIN:
		m.Cmdy(web.SPIDE, mdb.INPUTS, arg)
		m.Push(arg[0], "https://repos.shylinux.com")
	case OAUTH_URL:
		m.Push(arg[0], "/login/oauth/authorize")
	case GRANT_URL:
		m.Push(arg[0], "/login/oauth/access_token")
	case TOKEN_URL:
		m.Push(arg[0], "/login/oauth/access_token")
	case USERS_URL:
		m.Push(arg[0], "/api/v1/user")
	case lex.PREFIX:
		m.Push(arg[0], "token")
	case API:
		m.Push(arg[0], "/api/v1/")
	}
}
func (s Client) Auth(m *ice.Message, arg ...string) {
	m.Options(REDIRECT_URI, s.RedirectURI(m), RESPONSE_TYPE, CODE, STATE, m.Option(mdb.HASH))
	m.ProcessOpen(kit.MergeURL2(m.Option(web.DOMAIN), m.Option(OAUTH_URL), m.OptionSimple(CLIENT_ID, REDIRECT_URI, RESPONSE_TYPE, STATE)))
}
func (s Client) User(m *ice.Message, arg ...string) {
	if res := s.Get(m, m.Option(mdb.HASH), m.Option(USERS_URL)); res != nil {
		m.Options(res).Cmd(aaa.USER, mdb.CREATE, aaa.USERNICK, m.Option("full_name"), m.OptionSimple(aaa.USERNAME), aaa.AVATAR, m.Option(aaa.AVATAR_URL),
			aaa.USERROLE, kit.Select(aaa.VOID, aaa.TECH, m.Option("is_admin") == ice.TRUE), aaa.USERZONE, m.Option(web.DOMAIN),
			m.OptionSimple(aaa.EMAIL, aaa.LANGUAGE, aaa.AVATAR_URL))
	}
}
func (s Client) Orgs(m *ice.Message, arg ...string) {}
func (s Client) Repo(m *ice.Message, arg ...string) {}
func (s Client) List(m *ice.Message, arg ...string) {
	if s.Hash.List(m, arg...).PushAction(s.User, s.Auth, s.Remove); len(arg) == 0 {
		m.EchoScript(s.RedirectURI(m))
	} else {
		m.EchoScript("config header sso " + kit.MergeURL2(m.Append(web.DOMAIN), m.Append(OAUTH_URL), m.AppendSimple(CLIENT_ID), REDIRECT_URI, s.RedirectURI(m), RESPONSE_TYPE, CODE, STATE, arg[0]))
	}
}

func init() { ice.ChatCtxCmd(Client{}) }

func (s Client) RedirectURI(m *ice.Message) string {
	return strings.Split(web.MergeURL2(m, web.ChatCmdPath(m.Message, m.PrefixKey(), ctx.ACTION, aaa.LOGIN)), "?")[0]
}
func (s Client) Login(m *ice.Message, arg ...string) {
	if state, code := m.Option(STATE), m.Option(CODE); !m.Warn(state == "" || code == "") {
		s.Hash.List(m.Spawn(), m.Option(mdb.HASH, state)).Table(func(value ice.Maps) { m.Options(value) })
		m.Options(REDIRECT_URI, s.RedirectURI(m), GRANT_TYPE, AUTHORIZATION_CODE)
		if res := s.Post(m, m.Option(mdb.HASH), m.Option(GRANT_URL), m.OptionSimple(CLIENT_ID, CLIENT_SECRET, REDIRECT_URI, CODE, GRANT_TYPE)...); !m.Warn(res == nil) {
			kit.Value(res, EXPIRES_IN, m.Time(kit.Format("%vs", kit.Value(res, EXPIRES_IN))))
			s.Modify(m, kit.Simple(res)...)
			m.Options(res)
			if s.User(m); !m.Warn(m.Option(aaa.USERNAME) == "") && m.R != nil {
				web.RenderCookie(m.Message, aaa.SessCreate(m.Message, m.Option(aaa.USERNAME)))
				m.ProcessHistory()
			} else {
				m.ProcessClose()
			}
		}
	}
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
func (s Client) request(m *ice.Message, hash, api string, arg ...string) []string {
	msg := s.Hash.List(m.Spawn(), hash)
	kit.If(msg.Append(ACCESS_TOKEN), func(p string) {
		m.Options(web.SPIDE_HEADER, ice.Maps{web.Authorization: msg.Append(lex.PREFIX) + lex.SP + p})
	})
	kit.If(api == "", func() { api = path.Join(msg.Append(API), strings.ToLower(kit.FuncName(6))) })
	return kit.Simple(kit.MergeURL2(msg.Append(web.DOMAIN), api), arg)
}
