package lark

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
		web.P(SSO): {Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_USERNAME) != "" {
				web.RenderMain(m)
				return
			}
			appid := m.Cmd(APP).Append(APPID)
			home := m.MergeLink("/chat/lark/sso")
			if m.Option(cli.CODE) == "" {
				m.RenderRedirect(kit.MergeURL2(m.Cmd(web.SPIDE, LARK).Append("client.url"), "/open-apis/authen/v1/index"),
					"redirect_uri", kit.MergeURL(home, cli.BACK, m.R.Header.Get("Referer")), APP_ID, appid)
				return
			}
			msg := m.Cmd(web.SPIDE, LARK, "/open-apis/authen/v1/access_token", "grant_type", "authorization_code",
				cli.CODE, m.Option(cli.CODE), "app_access_token", m.Cmdx(APP, TOKEN, appid))
			msg = m.Cmd(EMPLOYEE, appid, m.Option(aaa.USERNAME, msg.Append("data.open_id")))
			m.Cmd(aaa.USER, mdb.CREATE, kit.Select(aaa.VOID, aaa.TECH, msg.Append("is_tenant_manager") == ice.TRUE), m.Option(aaa.USERNAME), "", "", LARK)
			m.Cmd(aaa.USER, mdb.MODIFY, aaa.AVATAR, msg.Append("avatar_url"), aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"),
				msg.AppendSimple(aaa.MOBILE, aaa.EMAIL, aaa.CITY, aaa.COUNTRY))
			m.RenderRedirect(m.MergeLink(kit.Select(home, m.Option(cli.BACK)), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(aaa.USERNAME))))
		}},
	})
}
