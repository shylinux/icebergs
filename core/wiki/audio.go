package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
)

const (
	M4A = "m4a"
)
const AUDIO = "audio"

func init() {
	Index.MergeCommands(ice.Commands{
		AUDIO: {Name: "audio url", Help: "音频", Actions: ctx.ConfAction(nfs.PATH, ice.USR_LOCAL_IMAGE), Hand: func(m *ice.Message, arg ...string) {
			_image_show(m, arg[0], arg[1:]...)
		}},
	})
}
