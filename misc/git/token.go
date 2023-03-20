package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TOKEN = "token"

func init() {
	const (
		SET = "set"
		GET = "get"
		SID = "sid"
	)
	const FILE = ".git-credentials"
	Index.MergeCommands(ice.Commands{
		TOKEN: {Name: "token username auto prunes", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, kit.Keys(TOKEN, SID)) }},
			web.PP(SET): {Hand: func(m *ice.Message, arg ...string) {
				host, list := kit.Dict(), []string{m.Option(TOKEN)}
				m.Cmd(nfs.CAT, kit.HomePath(FILE), func(line string) {
					kit.If(line != list[0], func() { kit.IfNoKey(host, kit.ParseURL(line).Host, func(p string) { list = append(list, line) }) })
				})
				m.Cmd(nfs.SAVE, kit.HomePath(FILE), strings.Join(list, ice.NL)+ice.NL)
				web.RenderTemplate(m, "close.html")
			}},
			web.PP(GET): {Hand: func(m *ice.Message, arg ...string) {
				m.W.Header().Set("Access-Control-Allow-Origin", "*")
				m.Cmd(nfs.CAT, kit.HomePath(FILE), func(text string) {
					if u := kit.ParseURL(text); u.Host == arg[0] {
						if p, ok := u.User.Password(); ok {
							m.Echo(u.User.Username()).Echo(p)
							m.W.Header().Set("Access-Control-Allow-Origin", u.Scheme+"://"+u.Host)
						}
					}
				})
			}},
			web.PP(SID): {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 1 && m.Cmd(TOKEN, arg[0]).Append(TOKEN) == arg[1] {
					web.RenderCookie(m.Echo(ice.OK), aaa.SessCreate(m, arg[0]))
				}
			}},
		}, mdb.HashAction(mdb.EXPIRE, mdb.MONTH, mdb.SHORT, aaa.USERNAME, mdb.FIELD, "time,username,token")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 && m.Length() > 0 {
				p := strings.Replace(m.Option(ice.MSG_USERHOST), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), m.Append(TOKEN)), 1)
				if m.EchoScript(p).EchoScript(nfs.Template(m, "echo.sh", p)); m.Option(ice.CMD) == m.PrefixKey() {
					m.ProcessReplace(kit.MergeURL2(m.Option(tcp.HOST), m.PrefixPath(SET), TOKEN, p))
				} else {
					m.EchoAnchor(kit.MergeURL2(m.Option(tcp.HOST), m.PrefixPath(SET), TOKEN, p))
				}
			}
		}},
	})
}
