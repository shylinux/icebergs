package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
)

const GRANT = "grant"

func init() {
	const CONFIRM = "confirm"
	Index.MergeCommands(ice.Commands{
		GRANT: {Name: "grant space auto", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			CONFIRM: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
					return
				}
				if m.Warn(m.Option(web.SPACE) == "", ice.ErrNotValid, web.SPACE) {
					return
				}
				if m.Warn(m.CmdAppend(web.SPACE, m.Option(web.SPACE), ice.CMD) != cli.PWD, ice.ErrNotFound, m.Option(web.SPACE)) {
					return
				}
				m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
				m.ProcessLocation(web.MergeURL2(m, ice.PS))
			}},
		}, aaa.RoleAction(CONFIRM)), Hand: func(m *ice.Message, arg ...string) {
			m.Echo("请授权: %s 访问设备: %s", arg[0], ice.Info.HostName).Echo(ice.NL).EchoButton(CONFIRM)
		}},
	})
}
