package macos

import (
	ice "shylinux.com/x/icebergs"
)

const CLOCK = "clock"

func init() {
	Index.MergeCommands(ice.Commands{
		CLOCK: {Icon: "Clock.png", Hand: func(m *ice.Message, arg ...string) { m.Display("") }},
	})
}
