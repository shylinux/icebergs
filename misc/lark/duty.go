package lark

import (
	ice "shylinux.com/x/icebergs"
)

const DUTY = "duty"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		DUTY: {Name: "duty [title] text run", Help: "通告", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			msg := m.Cmd(APP, m.Option(APP_ID))
			m.Cmdy(SEND, msg.Append(APPID), msg.Append(DUTY), arg)
		}},
	}})
}
