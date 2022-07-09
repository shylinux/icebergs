package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const SSO = "sso"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		"/sso": {Name: "/sso", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_USERNAME) == "" {
				m.RenderIndex(web.SERVE, ice.VOLCANOS)
				return
			}
			sessid := m.Cmdx(web.SPACE, m.Option(web.SPACE), aaa.SESS, mdb.CREATE,
				mdb.TIME, m.Time("720h"),
				aaa.USERROLE, m.Option(ice.MSG_USERROLE),
				aaa.USERNAME, m.Option(ice.MSG_USERNAME),
				aaa.USERNICK, m.Option(ice.MSG_USERNICK),
			)
			m.Cmdy(GRANT, mdb.INSERT, web.SPACE, m.Option(web.SPACE),
				aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK))
			m.RenderRedirect(kit.MergeURL(m.Option(cli.BACK), ice.MSG_SESSID, sessid))
			// web.RenderCookie(m, sessid, web.CookieName(m.Option("back")))
			// m.RenderRedirect(kit.MergeURL(m.Option("back")))
		}},
	}})
}
