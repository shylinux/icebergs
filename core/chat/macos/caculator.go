package macos

import (
	ice "shylinux.com/x/icebergs"
)

const CACULATOR = "caculator"

func init() {
	Index.MergeCommands(ice.Commands{
		CACULATOR: {Help: "计算器", Icon: "Caculator.png", Hand: func(m *ice.Message, arg ...string) { m.Display("") }},
	})
}
