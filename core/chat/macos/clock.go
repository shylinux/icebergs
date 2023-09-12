package macos

import (
	ice "shylinux.com/x/icebergs"
)

func init() {
	const CLOCK = "clock"
	Index.MergeCommands(ice.Commands{
		CLOCK: {Name: "clock refresh", Icon: "usr/icons/Clock.png", Hand: func(m *ice.Message, arg ...string) { m.Display("") }},
	})
}
