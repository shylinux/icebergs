package lark

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const SSO = "sso"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/sso": {Name: "/sso", Help: "网页", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if m.Option(ice.MSG_USERNAME) != "" { // 默认主页
				m.RenderIndex(web.SERVE, ice.VOLCANOS)
				return
			}

			home := kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/chat/lark/sso")
			if m.Option(kit.MDB_CODE) != "" { // 登录成功
				msg := m.Cmd(web.SPIDE, LARK, "/open-apis/authen/v1/access_token", "grant_type", "authorization_code",
					kit.MDB_CODE, m.Option(kit.MDB_CODE), "app_access_token", m.Cmdx(APP, TOKEN, m.Cmd(APP).Append(APPID)))

				// 创建会话
				m.Option(aaa.USERNAME, msg.Append("data.open_id"))
				web.RenderCookie(m, aaa.SessCreate(m, m.Option(aaa.USERNAME)))
				m.RenderRedirect(kit.Select(home, m.Option(kit.MDB_BACK)))

				// 更新用户
				msg = m.Cmd(EMPLOYEE, m.Option(aaa.USERNAME))
				m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERZONE, LARK, aaa.USERNICK, msg.Append(kit.MDB_NAME),
					aaa.AVATAR, msg.Append("avatar_url"), aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"),
					aaa.COUNTRY, msg.Append(aaa.COUNTRY), aaa.CITY, msg.Append(aaa.CITY),
					aaa.MOBILE, msg.Append(aaa.MOBILE),
				)
				return
			}

			if back := m.R.Header.Get("Referer"); back != "" {
				home = kit.MergeURL(home, kit.MDB_BACK, back)
			}
			// 登录页面
			m.RenderRedirect(kit.MergeURL2(m.Conf(APP, kit.Keym(LARK)), "/open-apis/authen/v1/index"),
				"redirect_uri", home, APP_ID, m.Cmd(APP).Append(APPID))
		}},
	}})
}
