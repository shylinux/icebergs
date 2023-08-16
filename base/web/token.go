package web

import (
	"encoding/base64"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
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
		TOKEN: {Name: "token username auto prunes", Help: "令牌", Actions: ice.MergeActions(ice.Actions{
			GEN: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo("请授权 %s 代码权限\n", m.Option(tcp.HOST)).EchoButton(CONFIRM)
			}},
			CONFIRM: {Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd("", m.Option(ice.MSG_USERNAME))
				if msg.Append(mdb.TIME) < m.Time() {
					msg = m.Cmd("", mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), TOKEN, kit.Hashs(mdb.UNIQ)).Cmd("", m.Option(ice.MSG_USERNAME))
				}
				m.ProcessReplace(kit.MergeURL2(m.Option(tcp.HOST), ChatCmdPath(m.PrefixKey(), SET), TOKEN, strings.Replace(UserHost(m), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), msg.Append(TOKEN)), 1)))
			}},
			SET: {Hand: func(m *ice.Message, arg ...string) {
				host, list := ice.Map{kit.ParseURL(m.Option(TOKEN)).Host: true}, []string{m.Option(TOKEN)}
				m.Cmd(nfs.CAT, kit.HomePath(FILE), func(line string) {
					line = strings.ReplaceAll(line, "%3a", ":")
					kit.IfNoKey(host, kit.ParseURL(line).Host, func(p string) { list = append(list, line) })
				}).Cmd(nfs.SAVE, kit.HomePath(FILE), strings.Join(list, lex.NL)+lex.NL)
				m.ProcessClose()
			}},
		}, mdb.HashAction(mdb.EXPIRE, mdb.MONTH, mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				return
				u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
				p := tcp.PublishLocalhost(m, kit.Format("%s://%s:%s@%s", u.Scheme, m.Append(aaa.USERNAME), m.Append(TOKEN), u.Host))
				m.EchoScript(p).EchoScript(kit.Format("echo '%s' >>~/.git-credentials", p))
			}
		}},
	})
	Index.MergeCommands(ice.Commands{
		"/check": {Hand: func(m *ice.Message, arg ...string) {
			kit.For(m.R.Header, func(key string, value []string) { m.Debug("what %v %v", key, value) })
			if BasicSess(m); m.Option(ice.MSG_USERNAME) == "" {
				BasicCheck(m, "请输入账号密码")
			}
		}},
		"/login": {Hand: func(m *ice.Message, arg ...string) { RenderMain(m) }},
		"/auths": {Hand: func(m *ice.Message, arg ...string) {
			kit.If(m.R.URL.Query().Get(ice.MSG_SESSID), func(p string) { RenderCookie(m, m.Option(ice.MSG_SESSID, p)) })
			RenderRedirect(m, m.R.URL.Query().Get("redirect_uri"))
		}},
	})
}

func BasicSess(m *ice.Message) {
	m.Options(ice.MSG_USERWEB, _serve_domain(m))
	m.Options(ice.MSG_SESSID, kit.Select(m.Option(ice.MSG_SESSID), m.Option(CookieName(m.Option(ice.MSG_USERWEB)))))
	aaa.SessCheck(m, m.Option(ice.MSG_SESSID))
}
func BasicCheck(m *ice.Message, realm string) bool {
	switch ls := kit.Split(m.R.Header.Get(Authorization)); kit.Select("", ls, 0) {
	case Basic:
		if buf, err := base64.StdEncoding.DecodeString(kit.Select("", ls, 1)); !m.Warn(err) {
			if ls := strings.SplitN(string(buf), ":", 2); !m.Warn(len(ls) < 2) {
				if msg := m.Cmd(TOKEN, ls[1]); !m.Warn(msg.Time() > msg.Append(mdb.TIME)) {
					return true
				}
			}
		}
	}
	m.W.Header().Add("WWW-Authenticate", kit.Format(`Basic realm="%s"`, realm))
	m.RenderStatusUnauthorized()
	return false
}
