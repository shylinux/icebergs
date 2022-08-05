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
		"/sso": {Name: "/sso", Help: "网页", Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_USERNAME) != "" { // 默认主页
				web.RenderIndex(m, ice.VOLCANOS)
				return
			}

			appid := m.Cmd(APP).Append(APPID)
			home := web.MergeURL2(m, "/chat/lark/sso")
			if m.Option(cli.CODE) == "" { // 登录页面
				if back := m.R.Header.Get("Referer"); back != "" {
					home = kit.MergeURL(home, cli.BACK, back)
				}
				m.RenderRedirect(kit.MergeURL2(m.Cmd(web.SPIDE, LARK).Append("client.url"), "/open-apis/authen/v1/index"),
					"redirect_uri", home, APP_ID, m.Cmd(APP).Append(APPID))
				return
			}

			msg := m.Cmd(web.SPIDE, LARK, "/open-apis/authen/v1/access_token", "grant_type", "authorization_code",
				cli.CODE, m.Option(cli.CODE), "app_access_token", m.Cmdx(APP, TOKEN, appid))

			// 更新用户
			m.Option(aaa.USERNAME, msg.Append("data.open_id"))
			msg = m.Cmd(EMPLOYEE, appid, m.Option(aaa.USERNAME))
			userrole := kit.Select(aaa.VOID, aaa.TECH, msg.Append("is_tenant_manager") == ice.TRUE)
			m.Cmd(aaa.USER, mdb.CREATE, m.Option(aaa.USERNAME), "", userrole)
			m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERROLE, userrole,
				aaa.USERNICK, msg.Append(mdb.NAME), aaa.USERZONE, LARK,
				aaa.AVATAR, msg.Append("avatar_url"), aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"),
				msg.AppendSimple(aaa.MOBILE, aaa.EMAIL, aaa.CITY, aaa.COUNTRY),
			)

			// 创建会话
			// web.RenderCookie(m, aaa.SessCreate(m, m.Option(aaa.USERNAME)), web.CookieName(m.Option(cli.BACK)))
			m.RenderRedirect(kit.MergeURL(kit.Select(home, m.Option(cli.BACK)), "sessid", aaa.SessCreate(m, m.Option(aaa.USERNAME))))
		}},
	})
}
