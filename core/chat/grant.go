package chat

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
)

const GRANT = "grant"

func init() {
	Index.MergeCommands(ice.Commands{
		GRANT: {Name: "grant space auto", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			web.SPACE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					link := tcp.PublishLocalhost(m, m.MergePodCmd("", "", web.SPACE, m.Option(mdb.NAME)))
					m.Sleep30ms(web.SPACE, m.Option(mdb.NAME), cli.PWD, m.Option(mdb.NAME), link, m.Cmdx(cli.QRCODE, link))
				})
			}},
			aaa.CONFIRM: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.R.Method == http.MethodGet, ice.ErrNotAllow) {
					return
				} else if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) || m.Warn(m.Option(web.SPACE) == "", ice.ErrNotValid, web.SPACE) {
					return
				} else if msg := m.Cmd(web.SPACE, m.Option(web.SPACE)); m.Warn(msg.Append(mdb.TYPE) != aaa.LOGIN, ice.ErrNotFound, m.Option(web.SPACE)) {
					return
				} else {
					m.Option(ice.MSG_USERUA, msg.Append(ice.MSG_USERUA))
					m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
					m.ProcessLocation(web.MergeURL2(m, msg.Append(mdb.TEXT)))
				}
			}},
		}, aaa.RoleAction(aaa.CONFIRM), gdb.EventAction(web.SPACE_LOGIN)), Hand: func(m *ice.Message, arg ...string) {
			m.Echo("请授权: %s 访问设备: %s", arg[0], ice.Info.Hostname).Echo(lex.NL).EchoButton(aaa.CONFIRM)
		}},
	})
}
