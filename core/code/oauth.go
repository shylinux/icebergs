package code

import (
	ice "shylinux.com/x/icebergs"
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
				m.Config(CLIENT_ID, m.Option(CLIENT_ID))
				m.Config(CLIENT_SECRET, m.Option(CLIENT_SECRET))
				m.Config(REDIRECT_URI, m.Option(REDIRECT_URI))
			}},
			"user": {Name: "user", Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, "Accept", web.ContentJSON, "Authorization", "token "+m.Option(ACCESS_TOKEN))
				data := kit.UnMarshal(m.Cmdx(web.SPIDE_GET, "https://api.github.com/user"))
				m.Cmdy(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH), kit.Simple(data))
			}},
			ACCESS_TOKEN: {Name: "access_token", Help: "访问", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, "Accept", web.ContentJSON)
				data := kit.UnMarshal(m.Cmdx(web.SPIDE_POST, kit.MergeURL2(LOGIN_OAUTH, ACCESS_TOKEN), m.ConfigSimple(CLIENT_ID, CLIENT_SECRET), m.OptionSimple(CODE)))
				m.Cmdy(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH), kit.Simple(data))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction("user", ACCESS_TOKEN, mdb.REMOVE); len(arg) == 0 {
				m.EchoAnchor(kit.MergeURL2(LOGIN_OAUTH, "authorize", m.ConfigSimple(REDIRECT_URI, CLIENT_ID)))
			}
		}},
		"/oauth": {Name: "/oauth", Help: "授权", Action: ice.MergeAction(map[string]*ice.Action{}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option(CODE) != "" {
				m.RenderCmd(m.PrefixKey(), m.Cmdx(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(CODE)))
			}
		}},
	}})
}
