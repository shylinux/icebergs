package wiki

import (
	ice "shylinux.com/x/icebergs"
)

const VIDEO = "video"

func init() {
	Index.MergeCommands(ice.Commands{
		VIDEO: {Name: "video path", Help: "视频", Hand: func(m *ice.Message, arg ...string) {
			arg = _name(m, arg)
			_image_show(m, arg[0], arg[1], arg[2:]...)
		}},
	})
}
