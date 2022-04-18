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
		CODE          = "code"
		ACCESS_TOKEN  = "access_token"
		LOGIN_OAUTH   = "https://github.com/login/oauth/"
	)
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		OAUTH: {Name: OAUTH, Help: "授权", Value: kit.Data(mdb.FIELD, "time,hash,code,access_token,scope,token_type")},
	}, Commands: map[string]*ice.Command{
		OAUTH: {Name: "oauth hash auto", Help: "授权", Action: ice.MergeAction(map[string]*ice.Action{
			"config": {Name: "config client_id client_secret redirect_uri", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				m.ConfigOption(CLIENT_ID, CLIENT_SECRET, REDIRECT_URI)
			}},
			"user": {Name: "user", Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, "Accept", web.ContentJSON, "Authorization", "token "+m.Option(ACCESS_TOKEN))
				data := kit.UnMarshal(m.Cmdx(web.SPIDE_GET, "https://api.github.com/user"))
				m.Cmdy(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH), kit.Simple(data))
			}},
			mdb.CREATE: {Name: "create code", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(CODE))
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.HASH {
					m.Cmdy(aaa.RSA).Cut("hash,title,public")
					return
				}
				m.Cmdy(mdb.INPUTS, m.PrefixKey(), "", mdb.HASH, arg)
			}},
			ACCESS_TOKEN: {Name: "access_token", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, "Accept", web.ContentJSON)
				data := kit.UnMarshal(m.Cmdx(web.SPIDE_POST, kit.MergeURL2(LOGIN_OAUTH, ACCESS_TOKEN), m.ConfigSimple(CLIENT_ID, CLIENT_SECRET), m.OptionSimple(CODE)))
				m.Cmdy(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH), kit.Simple(data))
			}},
			"public": {Name: "public hash", Help: "公钥", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, "Accept", web.ContentJSON, "Authorization", "token "+m.Option(ACCESS_TOKEN))
				msg := m.Cmd("aaa.rsa", m.Option(mdb.HASH))
				res := kit.UnMarshal(m.Cmdx(web.SPIDE_POST, kit.MergeURL2("https://api.github.com", "/user/keys"), web.SPIDE_JSON,
					"key", msg.Append("public"), msg.AppendSimple("title")))
				m.Push("", res)
				m.Echo("https://github.com/settings/keys")
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction("user", "public", ACCESS_TOKEN, mdb.REMOVE); len(arg) == 0 {
				m.Action(mdb.CREATE)
				m.Echo(kit.MergeURL2(LOGIN_OAUTH, "authorize", m.ConfigSimple(REDIRECT_URI, CLIENT_ID), "scope", "read:user read:public_key write:public_key repo"))
			}
		}},
		"/oauth": {Name: "/oauth", Help: "授权", Action: ice.MergeAction(map[string]*ice.Action{}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option(CODE) != "" {
				m.RenderCmd(m.PrefixKey(), m.Cmdx(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(CODE)))
			}
		}},
	}})
}
