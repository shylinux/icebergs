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
		OAUTH: {Name: "oauth hash auto", Help: "授权", Actions: ice.MergeAction(ice.Actions{
			ctx.CONFIG: {Name: "config client_id client_secret redirect_uri", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				for _, k := range []string{CLIENT_ID, CLIENT_SECRET, REDIRECT_URI} {
					m.Config(k, kit.Select(m.Config(k), m.Option(k)))
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.HASH {
					m.Cmdy(aaa.RSA).Cut("hash,title,public")
					return
				}
				m.Cmdy(mdb.INPUTS, m.PrefixKey(), "", mdb.HASH, arg)
			}},
			mdb.CREATE: {Name: "create code", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(CODE))
			}},
			ACCESS_TOKEN: {Name: "access_token", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON)
				data := kit.UnMarshal(m.Cmdx(web.SPIDE_POST, kit.MergeURL2(LOGIN_OAUTH, ACCESS_TOKEN), m.ConfigSimple(CLIENT_ID, CLIENT_SECRET), m.OptionSimple(CODE)))
				m.Cmdy(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH), kit.Simple(data))
			}},
			"public": {Name: "public hash", Help: "公钥", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+m.Option(ACCESS_TOKEN))
				msg := m.Cmd(aaa.RSA, m.Option(mdb.HASH))
				m.PushDetail(m.Cmdx(web.SPIDE_POST, API_GITHUB+"user/keys", web.SPIDE_JSON, "key", msg.Append("public"), msg.AppendSimple("title")))
				m.Echo("https://github.com/settings/keys")
			}},
			"user": {Name: "user", Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+m.Option(ACCESS_TOKEN))
				data := kit.UnMarshal(m.Cmdx(web.SPIDE_GET, API_GITHUB+"user"))
				m.Cmdy(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH), kit.Simple(data))
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m, m.Option(mdb.HASH))
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+m.Append(ACCESS_TOKEN))
				m.Cmd(web.SPIDE_DELETE, API_GITHUB+"user/keys/"+m.Option(mdb.ID))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,code,access_token,scope,token_type")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction("user", "public", ACCESS_TOKEN, mdb.REMOVE); len(arg) == 0 {
				m.Action(mdb.CREATE)
				m.Echo(kit.MergeURL2(LOGIN_OAUTH, "authorize", m.ConfigSimple(REDIRECT_URI, CLIENT_ID), SCOPE, "read:user read:public_key write:public_key repo"))
			} else if len(arg) == 1 {
				m.Option(web.SPIDE_HEADER, web.Accept, web.ContentJSON, web.Authorization, "token "+m.Append(ACCESS_TOKEN))
				m.SetAppend()
				m.Debug("what %v", m.FormatMeta())
				data := kit.UnMarshal(m.Cmdx(web.SPIDE_GET, API_GITHUB+"user/keys"))
				m.Debug("what %v", data)
				kit.Fetch(data, func(index int, value ice.Map) {
					m.PushRecord(value, "created_at,title,id,key")
				})
				m.PushAction(mdb.DELETE)
				m.Debug("what %v", m.FormatMeta())
			}
		}},
		"/oauth": {Name: "/oauth", Help: "授权", Actions: ice.MergeAction(ice.Actions{}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(CODE) != "" {
				web.RenderCmd(m, m.PrefixKey(), m.Cmdx(m.PrefixKey(), mdb.CREATE, m.OptionSimple(CODE)))
			}
		}},
	})
}
