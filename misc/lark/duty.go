package lark

import (
	ice "github.com/shylinux/icebergs"
)

const DUTY = "duty"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		DUTY: {Name: "duty appid [title] text auto", Help: "通告", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy(APP)
				return
			}

			duty := m.Cmd(APP, arg[0]).Append(DUTY)
			m.Cmdy(SEND, arg[0], duty, arg[1:])
		}},
	}})
}
