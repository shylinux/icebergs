package web

import (
	"net/http"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const TOKEN = "token"

func init() {
	const (
		GEN     = "gen"
		SET     = "set"
		CONFIRM = "confirm"
		FILE    = ".git-credentials"
		LOCAL   = "http://localhost:9020"
	)
	Index.MergeCommands(ice.Commands{
		TOKEN: {Help: "令牌", Actions: ice.MergeActions(ice.Actions{
			GEN: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoInfoButton(kit.Format("请授权 %s\n访问 %s\n", m.Option(tcp.HOST), m.Option(mdb.TYPE)), CONFIRM)
			}},
			CONFIRM: {Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.R.Method != http.MethodPost, ice.ErrNotAllow) {
					return
				}
				msg := m.Cmd("", mdb.CREATE, mdb.TYPE, m.Option(mdb.TYPE), mdb.NAME, m.Option(ice.MSG_USERNAME), mdb.TEXT, m.Option(tcp.HOST))
				m.ProcessReplace(kit.MergeURL2(m.Option(tcp.HOST), C(m.PrefixKey()), ctx.ACTION, SET,
					TOKEN, strings.Replace(UserHost(m), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), msg.Result()), 1)))
			}},
			SET: {Hand: func(m *ice.Message, arg ...string) {
				host, list := ice.Map{kit.ParseURL(m.Option(TOKEN)).Host: true}, []string{m.Option(TOKEN)}
				m.Cmd(nfs.CAT, kit.HomePath(FILE), func(line string) {
					line = strings.ReplaceAll(line, "%3a", ":")
					kit.IfNoKey(host, kit.ParseURL(line).Host, func(p string) { list = append(list, line) })
				}).Cmd(nfs.SAVE, kit.HomePath(FILE), strings.Join(list, lex.NL)+lex.NL)
				m.Cmd(cli.SYSTEM, "git", "config", "--global", "credential.helper", "store")
				m.ProcessClose()
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,type,name,text", mdb.EXPIRE, mdb.MONTH)), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.EchoScript(kit.Format("ish_miss_serve_log dev %s token %s", UserHost(m), arg[0]))
			}
		}},
	})
}
