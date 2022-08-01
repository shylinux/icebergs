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
	Index.MergeCommands(ice.Commands{
		"/sso": {Name: "/sso", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_USERNAME) == "" {
				m.RenderIndex(web.SERVE, ice.VOLCANOS)
				return
			}
			sessid := m.Cmdx(web.SPACE, m.Option(web.SPACE), aaa.SESS, mdb.CREATE,
				aaa.USERNAME, m.Option(ice.MSG_USERNAME),
				aaa.USERROLE, m.Option(ice.MSG_USERROLE),
				aaa.USERNICK, m.Option(ice.MSG_USERNICK),
			)
			m.RenderRedirect(kit.MergeURL(m.Option(cli.BACK), ice.MSG_SESSID, sessid))

			// m.Cmdy(GRANT, mdb.INSERT, web.SPACE, m.Option(web.SPACE),
			// 	aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK))
			// web.RenderCookie(m, sessid, web.CookieName(m.Option("back")))
			// m.RenderRedirect(kit.MergeURL(m.Option("back")))
		}},
	})
}
