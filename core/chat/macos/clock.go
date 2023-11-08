package macos

import (
	ice "shylinux.com/x/icebergs"
)

const CLOCK = "clock"

func init() {
	Index.MergeCommands(ice.Commands{
		CLOCK: {Help: "时钟", Icon: "Clock.png", Hand: func(m *ice.Message, arg ...string) { m.Display("") }},
	})
}
