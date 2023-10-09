package wiki

import (
	ice "shylinux.com/x/icebergs"
)

const AUDIO = "audio"

func init() {
	Index.MergeCommands(ice.Commands{
		AUDIO: {Name: "audio path", Help: "音频", Hand: func(m *ice.Message, arg ...string) {
			arg = _name(m, arg)
			_image_show(m, arg[0], arg[1], arg[2:]...)
		}},
	})
}
