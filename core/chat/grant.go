package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
)

const GRANT = "grant"

func init() {
	const CONFIRM = "confirm"
	Index.MergeCommands(ice.Commands{
		GRANT: {Name: "grant space auto", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			web.SPACE_OPEN: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TYPE) != aaa.LOGIN {
					return
				}
				m.Go(func(msg *ice.Message) {
					link := tcp.PublishLocalhost(m, web.MergePodCmd(m, "", "", web.SPACE, m.Option(mdb.NAME)))
					msg.Sleep300ms(web.SPACE, m.Option(mdb.NAME), cli.PWD, m.Option(mdb.NAME), link, msg.Cmdx(cli.QRCODE, link))
				})
			}},
			CONFIRM: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) || m.Warn(m.Option(web.SPACE) == "", ice.ErrNotValid, web.SPACE) {
					return
				}
				if msg := m.Cmd(web.SPACE, m.Option(web.SPACE)); m.Warn(msg.Append(mdb.TYPE) != aaa.LOGIN, ice.ErrNotFound, m.Option(web.SPACE)) {
					return
				} else {
					m.Option(ice.MSG_USERUA, msg.Append(ice.MSG_USERUA))
				}
				m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
				m.ProcessLocation(web.MergeURL2(m, ice.PS))
			}},
		}, gdb.EventAction(web.SPACE_OPEN), aaa.RoleAction(CONFIRM)), Hand: func(m *ice.Message, arg ...string) {
			m.Echo("请授权: %s 访问设备: %s", arg[0], ice.Info.HostName).Echo(ice.NL).EchoButton(CONFIRM)
		}},
	})
}
