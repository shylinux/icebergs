package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
)

func init() {
	const CLOCK = "clock"
	Index.MergeCommands(ice.Commands{
		CLOCK: {Name: "clock refresh", Icon: "usr/icons/Clock.png", Hand: func(m *ice.Message, arg ...string) {
			ctx.DisplayLocal(m, "")
		}},
	})
}
