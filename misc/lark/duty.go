package lark

import (
	ice "shylinux.com/x/icebergs"
)

const DUTY = "duty"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		DUTY: {Name: "duty [title] text auto", Help: "通告", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			duty := m.Cmd(APP, m.Option(APP_ID)).Append(DUTY)
			m.Cmdy(SEND, arg[0], duty, arg[1:])
		}},
	}})
}
