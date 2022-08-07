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
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON)
				data := web.SpidePost(m, kit.MergeURL2(LOGIN_OAUTH, ACCESS_TOKEN), m.ConfigSimple(CLIENT_ID, CLIENT_SECRET), m.OptionSimple(CODE))
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), kit.Simple(data))
			}},
			"public": {Name: "public hash", Help: "公钥", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+m.Option(ACCESS_TOKEN))
				msg := m.Cmd(aaa.RSA, m.Option(mdb.HASH))
				m.PushDetail(web.SpidePost(m, API_GITHUB+"user/keys", web.SPIDE_JSON, "key", msg.Append("public"), msg.AppendSimple("title")))
				m.Echo("https://github.com/settings/keys")
			}},
			"user": {Name: "user", Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+m.Option(ACCESS_TOKEN))
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), kit.Simple(web.SpideGet(m, API_GITHUB+"user")))
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m, m.Option(mdb.HASH))
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+m.Append(ACCESS_TOKEN))
				web.SpideDelete(m, API_GITHUB+"user/keys/"+m.Option(mdb.ID))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,code,access_token,scope,token_type")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction("user", "public", ACCESS_TOKEN, mdb.REMOVE); len(arg) == 0 {
				m.Echo(kit.MergeURL2(LOGIN_OAUTH, "authorize", m.ConfigSimple(REDIRECT_URI, CLIENT_ID), SCOPE, "read:user read:public_key write:public_key repo"))
				m.Action(mdb.CREATE)
			} else if len(arg) == 1 {
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+m.Append(ACCESS_TOKEN))
				m.SetAppend()
				kit.Fetch(web.SpideGet(m, API_GITHUB+"user/keys"), func(index int, value ice.Map) {
					m.PushRecord(value, "created_at,title,id,key")
				})
				m.PushAction(mdb.DELETE)
			}
		}},
		"/oauth": {Name: "/oauth", Help: "授权", Actions: ice.MergeActions(ice.Actions{}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(CODE) != "" {
				web.RenderCmd(m, m.PrefixKey(), m.Cmdx(m.PrefixKey(), mdb.CREATE, m.OptionSimple(CODE)))
			}
		}},
	})
}
