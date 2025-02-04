package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
)

const VIDEO = "video"

func init() {
	Index.MergeCommands(ice.Commands{
		VIDEO: {Name: "video path", Help: "视频", Actions: ice.Actions{
			"material": {Hand: func(m *ice.Message, arg ...string) {
				if nfs.Exists(m, nfs.USR_MATERIAL) {
					m.Cmdy(VIDEO, path.Join(nfs.USR_MATERIAL, strings.TrimPrefix(path.Dir(m.Option("_script")), nfs.USR), arg[0]))
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			arg = _name(m, arg)
			_image_show(m, arg[0], arg[1], arg[2:]...)
		}},
	})
}
