package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const GRANT = "grant"

func init() {
	Index.MergeCommands(ice.Commands{
		GRANT: {Name: "grant space id auto", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			"confirm": {Help: "通过", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
				m.ProcessLocation(web.MergeURL2(m, ice.PS))
			}},
		}, mdb.HashAction(mdb.SHORT, web.SPACE, mdb.FIELD, "time,space,userrole,username,usernick"), aaa.RoleAction("confirm")), Hand: func(m *ice.Message, arg ...string) {
			m.Echo("请授权: %s 访问设备: %s", arg[0], ice.Info.HostName).Echo(ice.NL)
			m.EchoButton("confirm")
		}},
	})
}
