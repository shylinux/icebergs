package web

import (
	"encoding/base64"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.MergeCommands(ice.Commands{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, "basic") }},
		"/basic/check": {Hand: func(m *ice.Message, arg ...string) {
			kit.For(m.R.Header, func(key string, value []string) { m.Debug("what %v %v", key, value) })
			if BasicSess(m); m.Option(ice.MSG_USERNAME) == "" {
				BasicCheck(m, "请输入账号密码")
			}
		}},
		"/basic/login": {Hand: func(m *ice.Message, arg ...string) { RenderMain(m) }},
		"/basic/auths": {Hand: func(m *ice.Message, arg ...string) {
			kit.If(m.R.URL.Query().Get(ice.MSG_SESSID), func(p string) { RenderCookie(m, m.Option(ice.MSG_SESSID, p)) })
			RenderRedirect(m, kit.Select(nfs.PS, m.R.URL.Query().Get("redirect_uri")))
		}},
	})
}
func BasicSess(m *ice.Message) {
	m.Options(ice.MSG_USERWEB, _serve_domain(m))
	m.Options(ice.MSG_SESSID, kit.Select(m.Option(ice.MSG_SESSID), m.Option(CookieName(m.Option(ice.MSG_USERWEB)))))
	aaa.SessCheck(m, m.Option(ice.MSG_SESSID))
}
func BasicCheck(m *ice.Message, realm string, check ...func(*ice.Message) bool) bool {
	switch ls := kit.Split(m.R.Header.Get(html.Authorization)); kit.Select("", ls, 0) {
	case html.Basic:
		if buf, err := base64.StdEncoding.DecodeString(kit.Select("", ls, 1)); !m.WarnNotValid(err) {
			if ls := strings.SplitN(string(buf), ":", 2); !m.WarnNotValid(len(ls) < 2 || ls[1] == "", html.Basic) {
				if msg := m.Cmd(TOKEN, ls[1]); !m.WarnNotValid(msg.Time() > msg.Append(mdb.TIME)) {
					if len(check) == 0 || check[0](msg) {
						return true
					}
				}
			}
		}
	}
	m.W.Header().Add("WWW-Authenticate", kit.Format(`Basic realm="%s"`, realm))
	m.RenderStatusUnauthorized()
	return false
}
