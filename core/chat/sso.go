package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const SSO = "sso"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/sso": {Name: "/sso", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			sessid := m.Cmdx(web.SPACE, m.Option(web.SPACE), aaa.SESS, mdb.CREATE,
				mdb.TIME, m.Time("720h"),
				aaa.USERROLE, m.Option(ice.MSG_USERROLE),
				aaa.USERNAME, m.Option(ice.MSG_USERNAME),
				aaa.USERNICK, m.Option(ice.MSG_USERNICK),
			)
			m.RenderRedirect(kit.MergeURL(m.Option("back"), ice.MSG_SESSID, sessid))
		}},
	}})
}
