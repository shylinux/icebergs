package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
)

const VIDEO = "video"

func init() {
	Index.MergeCommands(ice.Commands{
		VIDEO: {Name: "video path", Help: "视频", Actions: ice.Actions{
			"material": {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(VIDEO, path.Join("usr/material", strings.TrimPrefix(path.Dir(m.Option("_script")), "usr/"), arg[0]))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			arg = _name(m, arg)
			_image_show(m, arg[0], arg[1], arg[2:]...)
		}},
	})
}
