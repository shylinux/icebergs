package wiki

import (
	ice "shylinux.com/x/icebergs"
)
const AUDIO = "audio"

func init() {
	Index.MergeCommands(ice.Commands{
		AUDIO: {Name: "audio path auto", Help: "音频", Actions: ice.MergeActions(
		), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 {
				m.Echo("<audio class='story' src='%s'></audio>", arg[0])
			}
		}},
	})
}