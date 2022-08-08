package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const OAUTH = "oauth"

func init() {
	const (
		CLIENT_ID     = "client_id"
		CLIENT_SECRET = "client_secret"
		REDIRECT_URI  = "redirect_uri"
		SCOPE         = "scope"
		CODE          = "code"
		ACCESS_TOKEN  = "access_token"
		LOGIN_OAUTH   = "https://github.com/login/oauth/"
		API_GITHUB    = "https://api.github.com/"
	)
	_oauth_header := func(m *ice.Message, arg ...string) *ice.Message {
		m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+kit.Select(m.Option(ACCESS_TOKEN), arg, 0))
		return m
	}

	Index.MergeCommands(ice.Commands{
		OAUTH: {Name: "oauth hash auto", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			ctx.CONFIG: {Name: "config client_id client_secret redirect_uri", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				ctx.ConfigFromOption(m, CLIENT_ID, CLIENT_SECRET, REDIRECT_URI)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.HASH {
					m.Cmdy(aaa.RSA).Cut("hash,title,public")
					return
				}
				mdb.HashInputs(m, arg)
			}},
			mdb.CREATE: {Name: "create code", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(CODE))
			}},
			ACCESS_TOKEN: {Name: "access_token", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), kit.Simple(web.SpidePost(_oauth_header(m), kit.MergeURL2(LOGIN_OAUTH, ACCESS_TOKEN), m.ConfigSimple(CLIENT_ID, CLIENT_SECRET), m.OptionSimple(CODE))))
			}},
			"user": {Name: "user", Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), kit.Simple(web.SpideGet(_oauth_header(m), API_GITHUB+"user")))
			}},
			"public": {Name: "public hash", Help: "公钥", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(aaa.RSA, m.Option(mdb.HASH))
				m.PushDetail(web.SpidePost(_oauth_header(m), API_GITHUB+"user/keys", web.SPIDE_JSON, "key", msg.Append("public"), msg.AppendSimple("title")))
				m.Echo("https://github.com/settings/keys")
			}},
			"keys": {Name: "keys", Help: "用户密钥", Hand: func(m *ice.Message, arg ...string) {
				_oauth_header(m).SetAppend()
				kit.Fetch(web.SpideGet(m, API_GITHUB+"user/keys"), func(index int, value ice.Map) {
					m.PushRecord(value, "created_at,title,id,key")
				})
				m.PushAction(mdb.DELETE)
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m, m.Option(mdb.HASH))
				web.SpideDelete(_oauth_header(m), API_GITHUB+"user/keys/"+m.Option(mdb.ID))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,code,access_token,scope,token_type")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction("public", "user", ACCESS_TOKEN, mdb.REMOVE); len(arg) == 0 {
				if m.Action(mdb.CREATE); m.Length() == 0 {
					m.Echo(kit.MergeURL2(LOGIN_OAUTH, "authorize", m.ConfigSimple(REDIRECT_URI, CLIENT_ID), SCOPE, "read:user read:public_key write:public_key repo"))
				}
			}
		}},
		"/oauth": {Name: "/oauth", Help: "授权", Actions: ice.MergeActions(ice.Actions{}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(CODE) == "", ice.ErrNotValid) {
				return
			}
			web.RenderCmd(m, m.PrefixKey(), m.Cmdx(m.PrefixKey(), mdb.CREATE, m.OptionSimple(CODE)))
		}},
	})
}
