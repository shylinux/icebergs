package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const OAUTH = "oauth"

func init() {
	const (
		CHECK = "check"
		APPLY = "apply"
		REPLY = "reply"
		OFFER = "offer"

		OAUTH_APPLY = "/oauth/apply"
		OAUTH_REPLY = "/oauth/reply"
		OAUTH_OFFER = "/oauth/offer"
	)
	const (
		SCOPE        = "scope"
		REDIRECT_URI = "redirect_uri"
		ACCESS_TOKEN = "access_token"
		EXPIRES      = "expires"
	)
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		OAUTH: {Name: "oauth hash auto prunes", Help: "授权", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.Prefix(OAUTH_APPLY))
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.Prefix(OAUTH_OFFER))
			}},
			CHECK: {Name: "check scope", Help: "检查", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.MergeURL(m.Cmdx(cli.RUNTIME, cli.MAKE_DOMAIN)+OAUTH_APPLY, m.OptionSimple(SCOPE), REDIRECT_URI,
					kit.Select(m.MergePodURL(OAUTH_REPLY), m.MergeLink("/chat/"+OAUTH_REPLY), m.Option(ice.MSG_USERPOD) == "")))
			}},
			APPLY: {Name: "apply redirect_uri", Help: "申请", Hand: func(m *ice.Message, arg ...string) {
				if m.Right(m.Option(SCOPE)) {
					token := mdb.HashCreate(m, mdb.TIME, m.Time(m.Config(EXPIRES)), aaa.USERNAME, m.Option(ice.MSG_USERNAME), m.OptionSimple(SCOPE, REDIRECT_URI)).Result()
					m.ProcessReplace(kit.MergeURL(m.Option(REDIRECT_URI), OFFER, m.MergePodURL(OAUTH_OFFER, ACCESS_TOKEN, token)))
				}
			}},
			REPLY: {Name: "reply offer", Help: "通过", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.UserAgent, m.PrefixKey())
				m.Cmd(ssh.SOURCE, ice.ETC_LOCAL_SHY, kit.Dict(nfs.CAT_CONTENT, m.Cmdx(web.SPIDE, ice.DEV, web.SPIDE_GET, m.Option(OFFER))))
				m.ProcessHistory()
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.UNIQ, EXPIRES, "720h"), ctx.CmdAction())},
		OAUTH_APPLY: {Name: "/oauth/apply", Help: "授权申请", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.RenderCmd(m.Prefix(OAUTH), APPLY)
		}},
		OAUTH_REPLY: {Name: "/oauth/reply", Help: "授权通过", Action: ctx.CmdAction(), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.RenderCmd(m.Prefix(OAUTH), REPLY)
		}},
		OAUTH_OFFER: {Name: "/oauth/offer", Help: "授权资源", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if msg := m.Cmd(OAUTH, m.Option(ACCESS_TOKEN), ice.OptionFields("time,scope")); kit.Time(msg.Time()) < kit.Time(msg.Append(mdb.TIME)) {
				aaa.UserRoot(m).Cmdy(nfs.CAT, msg.Append(SCOPE)).RenderResult()
			}
		}},
	}})
}
