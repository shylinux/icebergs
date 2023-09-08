package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
)

func init() {
	const CACULATOR = "caculator"
	Index.MergeCommands(ice.Commands{
		CACULATOR: {Name: "caculator refresh", Icon: "usr/icons/Caculator.png", Hand: func(m *ice.Message, arg ...string) {
			ctx.DisplayLocal(m, "")
		}},
	})
}
