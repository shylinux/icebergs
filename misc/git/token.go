package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
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
		FILE  = ".git-credentials"
		LOCAL = "http://localhost:9020"
	)
	Index.MergeCommands(ice.Commands{
		TOKEN: {Name: "token username auto prunes", Help: "令牌", Actions: ice.MergeActions(ice.Actions{
			GEN: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo("请授权 %s 代码权限\n", m.Option(tcp.HOST)).EchoButton("confirm")
			}},
			"confirm": {Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd("", m.Option(ice.MSG_USERNAME))
				if msg.Append(mdb.TIME) < m.Time() {
					msg = m.Cmd("", mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), TOKEN, kit.Hashs(mdb.UNIQ)).Cmd("", m.Option(ice.MSG_USERNAME))
				}
				m.ProcessReplace(kit.MergeURL2(m.Option(tcp.HOST), web.ChatCmdPath(m.PrefixKey(), SET), TOKEN, strings.Replace(web.UserHost(m), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), msg.Append(TOKEN)), 1)))
			}},
			SET: {Hand: func(m *ice.Message, arg ...string) {
				host, list := ice.Map{kit.ParseURL(m.Option(TOKEN)).Host: true}, []string{m.Option(TOKEN)}
				m.Cmd(nfs.CAT, kit.HomePath(FILE), func(line string) {
					kit.IfNoKey(host, kit.ParseURL(line).Host, func(p string) { list = append(list, line) })
				}).Cmd(nfs.SAVE, kit.HomePath(FILE), strings.Join(list, lex.NL)+lex.NL)
				m.ProcessClose()
			}},
		}, mdb.HashAction(mdb.EXPIRE, mdb.MONTH, mdb.SHORT, aaa.USERNAME, mdb.FIELD, "time,username,token")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
				p := tcp.PublishLocalhost(m, kit.Format("%s://%s:%s@%s", u.Scheme, m.Append(aaa.USERNAME), m.Append(TOKEN), u.Host))
				m.EchoScript(p).EchoScript(kit.Format("echo '%s' >>~/.git-credentials", p))
			}
		}},
	})
}
