package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const SSO = "sso"

func init() {
	Index.MergeCommands(ice.Commands{
		SSO: {Actions: ice.MergeActions(web.ApiAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) || m.Warn(m.Option(web.SPACE) == "", ice.ErrNotValid) || m.Warn(m.Option(cli.BACK) == "", ice.ErrNotValid) {
				web.RenderMain(m)
				return
			}
			msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
			m.Cmdx(web.SPACE, m.Option(web.SPACE), aaa.USER, mdb.CREATE, msg.AppendSimple(aaa.USERNICK, aaa.USERNAME, aaa.USERROLE, aaa.USERZONE, aaa.AVATAR, aaa.AVATAR_URL, aaa.BACKGROUND, aaa.LANGUAGE))
			m.RenderRedirect(kit.MergeURL(m.Option(cli.BACK), ice.MSG_SESSID, m.Cmdx(web.SPACE, m.Option(web.SPACE), aaa.SESS, mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERROLE, m.Option(ice.MSG_USERROLE), kit.Dict(ice.MSG_USERROLE, aaa.TECH))))
		}},
	})
}

func GetSSO(m *ice.Message) string {
	link := m.Cmdx(web.SPACE, web.DOMAIN)
	if !strings.Contains(link, web.PP(CHAT, POD)) {
		return ""
	}
	ls := strings.Split(kit.ParseURL(link).Path, nfs.PS)
	return kit.MergeURL2(link, web.PP(CHAT, SSO), web.SPACE, kit.Select("", ls, 3), cli.BACK, m.R.Header.Get(web.Referer))
}
