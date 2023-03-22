package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const GRANT = "grant"

func init() {
	Index.MergeCommands(ice.Commands{
		GRANT: {Name: "grant hash auto", Help: "授权", Actions: ice.Actions{
			mdb.MODIFY: {Help: "同意", Hand: func(m *ice.Message, arg ...string) { m.Cmd(SESS, mdb.MODIFY, GRANT, m.Option(ice.MSG_USERNAME)) }},
			mdb.REVERT: {Help: "撤销", Hand: func(m *ice.Message, arg ...string) { m.Cmd(SESS, mdb.MODIFY, GRANT, "") }},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(SESS, arg); len(arg) > 0 && m.Append(GRANT) == "" {
				m.Echo("请授权 %s@%s 访问 %s", m.Append(aaa.USERNAME), m.Append(tcp.HOSTNAME), web.UserHost(m))
			}
			m.Tables(func(value ice.Maps) { m.PushButton(kit.Select(mdb.REVERT, mdb.MODIFY, value[GRANT] == "")) })
		}},
	})
}
