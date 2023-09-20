package macos

import (
	ice "shylinux.com/x/icebergs"
)

func init() {
	const CACULATOR = "caculator"
	Index.MergeCommands(ice.Commands{
		CACULATOR: {Name: "caculator refresh", Icon: "Caculator.png", Hand: func(m *ice.Message, arg ...string) { m.Display("") }},
	})
}
