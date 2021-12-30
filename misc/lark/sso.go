package lark

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
		"/sso": {Name: "/sso", Help: "网页", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if m.Option(ice.MSG_USERNAME) != "" { // 默认主页
				m.RenderIndex(web.SERVE, ice.VOLCANOS)
				return
			}

			appid := m.Cmd(APP).Append(APPID)
			home := m.MergeURL2("/chat/lark/sso")
			if m.Option(kit.MDB_CODE) != "" { // 登录成功
				msg := m.Cmd(web.SPIDE, LARK, "/open-apis/authen/v1/access_token", "grant_type", "authorization_code",
					kit.MDB_CODE, m.Option(kit.MDB_CODE), "app_access_token", m.Cmdx(APP, TOKEN, appid))

				// 更新用户
				m.Option(aaa.USERNAME, msg.Append("data.open_id"))
				msg = m.Cmd(EMPLOYEE, appid, m.Option(aaa.USERNAME))
				userrole := kit.Select(aaa.VOID, aaa.TECH, msg.Append("is_tenant_manager") == ice.TRUE)
				m.Cmd(aaa.USER, mdb.CREATE, userrole, m.Option(aaa.USERNAME))
				m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERROLE, userrole,
					aaa.USERNICK, msg.Append(kit.MDB_NAME), aaa.USERZONE, LARK,
					aaa.AVATAR, msg.Append("avatar_url"), aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"),
					msg.AppendSimple(aaa.MOBILE, aaa.EMAIL, aaa.CITY, aaa.COUNTRY),
				)

				// 创建会话
				web.RenderCookie(m, aaa.SessCreate(m, m.Option(aaa.USERNAME)), web.CookieName(m.Option(kit.MDB_BACK)))
				m.RenderRedirect(kit.Select(home, m.Option(kit.MDB_BACK)))
				return
			}

			if back := m.R.Header.Get("Referer"); back != "" {
				home = kit.MergeURL(home, kit.MDB_BACK, back)
			}
			// 登录页面
			m.RenderRedirect(kit.MergeURL2(m.Cmd(web.SPIDE, LARK).Append("client.url"), "/open-apis/authen/v1/index"),
				"redirect_uri", home, APP_ID, m.Cmd(APP).Append(APPID))
		}},
	}})
}
