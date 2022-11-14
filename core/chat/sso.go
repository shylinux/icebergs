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
		web.P(SSO): {Name: "/sso", Help: "授权", Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_USERNAME) == "" {
				web.RenderIndex(m, ice.VOLCANOS)
				return
			}
			if m.Warn(m.Option(cli.BACK) == "") {
				return
			}
			sessid := aaa.UserRoot(m).Cmdx(web.SPACE, m.Option(web.SPACE), aaa.SESS, mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME),
				aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNICK, m.Option(ice.MSG_USERNICK))
			m.RenderRedirect(kit.MergeURL(m.Option(cli.BACK), ice.MSG_SESSID, sessid))
		}},
	})
}
