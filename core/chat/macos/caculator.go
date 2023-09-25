package macos

import (
	ice "shylinux.com/x/icebergs"
)

const CACULATOR = "caculator"

func init() {
	Index.MergeCommands(ice.Commands{
		CACULATOR: {Icon: "Caculator.png", Hand: func(m *ice.Message, arg ...string) { m.Display("") }},
	})
}
