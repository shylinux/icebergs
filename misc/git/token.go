package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TOKEN = "token"

func init() {
	Index.MergeCommands(ice.Commands{
		TOKEN: {Name: "token username auto prunes", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, "token.sid") }},
			web.PP("get"): {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.CAT, kit.HomePath(".git-credentials"), func(text string) {
					if strings.HasSuffix(text, ice.AT+arg[0]) {
						u := kit.ParseURL(text)
						if p, ok := u.User.Password(); ok {
							m.Echo(u.User.Username()).Echo(p)
							m.W.Header().Add("Access-Control-Allow-Origin", u.Scheme+"://"+arg[0])
						}
					}
				})
			}},
			web.PP("sid"): {Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(TOKEN, arg[0]).Append(TOKEN) == arg[1] {
					web.RenderCookie(m, aaa.SessCreate(m, arg[0]))
					m.Echo(ice.OK)
				}
			}},
		}, mdb.HashAction(mdb.EXPIRE, mdb.MONTH, mdb.SHORT, aaa.USERNAME, mdb.FIELD, "time,username,token")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 && m.Length() > 0 {
				m.EchoScript(strings.Replace(m.Option(ice.MSG_USERHOST), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), m.Append(TOKEN)), 1))
				m.EchoScript(nfs.Template(m, "echo.sh", strings.Replace(m.Option(ice.MSG_USERHOST), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), m.Append(TOKEN)), 1)))
			}
		}},
	})
}
