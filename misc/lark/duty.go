package lark

import (
	ice "shylinux.com/x/icebergs"
)

const DUTY = "duty"

func init() {
	Index.MergeCommands(ice.Commands{
		DUTY: {Name: "duty [title] text run", Help: "通告", Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(APP, m.Option(APP_ID))
			m.Cmdy(SEND, msg.Append(APPID), msg.Append(DUTY), arg)
		}},
	})
}
