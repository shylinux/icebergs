package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TOKEN = "token"

func init() {
	const (
		GEN   = "gen"
		SET   = "set"
		GET   = "get"
		SID   = "sid"
		FILE  = ".git-credentials"
		LOCAL = "http://localhost:9020"
	)
	create := func(m *ice.Message) string {
		msg := m.Cmd(Prefix(TOKEN), m.Cmdx(Prefix(TOKEN), mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), TOKEN, kit.Hashs(mdb.UNIQ)))
		return strings.Replace(web.UserHost(m), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), msg.Append(TOKEN)), 1)
	}
	Index.MergeCommands(ice.Commands{
		TOKEN: {Name: "token username auto prunes", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, kit.Keys(TOKEN, SID)) }},
			cli.MAKE: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessReplace(kit.MergeURL2(LOCAL, m.PrefixPath(SET), TOKEN, create(m)))
			}},
			web.PP(GEN): {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessReplace(kit.MergeURL2(m.Option(tcp.HOST), m.PrefixPath(SET), TOKEN, create(m)))
			}},
			web.PP(SET): {Hand: func(m *ice.Message, arg ...string) {
				defer web.RenderTemplate(m, "close.html")
				host, list := ice.Map{kit.ParseURL(m.Option(TOKEN)).Host: true}, []string{m.Option(TOKEN)}
				m.Cmd(nfs.CAT, kit.HomePath(FILE), func(line string) {
					kit.IfNoKey(host, kit.ParseURL(line).Host, func(p string) { list = append(list, line) })
				}).Cmd(nfs.SAVE, kit.HomePath(FILE), strings.Join(list, ice.NL)+ice.NL)
			}},
			web.PP(GET): {Hand: func(m *ice.Message, arg ...string) {
				web.RenderOrigin(m.W, "*")
				m.Cmd(nfs.CAT, kit.HomePath(FILE), func(text string) {
					if u := kit.ParseURL(text); u.Host == arg[0] {
						if p, ok := u.User.Password(); ok {
							m.Echo(u.User.Username()).Echo(p)
							web.RenderOrigin(m.W, u.Scheme+"://"+u.Host)
						}
					}
				})
			}},
			web.PP(SID): {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 1 && m.Cmd(TOKEN, arg[0]).Append(TOKEN) == arg[1] {
					web.RenderCookie(m.Echo(ice.OK), aaa.SessCreate(m, arg[0]))
				}
			}},
		}, mdb.HashAction(mdb.EXPIRE, mdb.MONTH, mdb.SHORT, aaa.USERNAME, mdb.FIELD, "time,username,token"))},
	})
}
