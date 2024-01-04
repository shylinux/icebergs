package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const SSO = "sso"

func init() {
	Index.MergeCommands(ice.Commands{
		SSO: {Help: "授权", Actions: web.ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) || m.Warn(m.Option(web.SPACE) == "", ice.ErrNotValid) || m.Warn(m.Option(cli.BACK) == "", ice.ErrNotValid) {
				web.RenderMain(m)
				return
			}
			m.Cmdx(web.SPACE, m.Option(web.SPACE), aaa.USER, mdb.CREATE, m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME)).AppendSimple(aaa.USERNICK, aaa.USERNAME, aaa.USERROLE, aaa.USERZONE, aaa.AVATAR, aaa.AVATAR_URL, aaa.BACKGROUND, aaa.LANGUAGE))
			m.RenderRedirect(kit.MergeURL2(m.Option(cli.BACK), web.P(web.SHARE, m.Cmdx(web.SPACE, m.Option(web.SPACE), web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN, mdb.TEXT, m.Option(cli.BACK)))))
		}},
	})
}

func GetSSO(m *ice.Message) string {
	link := m.Cmdx(web.SPACE, web.DOMAIN)
	if !strings.Contains(link, web.S()) {
		return ""
	}
	ls := strings.Split(kit.ParseURL(link).Path, nfs.PS)
	return kit.MergeURL2(link, web.PP(CHAT, SSO), web.SPACE, kit.Select("", ls, 2), cli.BACK, m.R.Header.Get(html.Referer))
}
