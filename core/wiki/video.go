package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
)

const (
	m4v = "m4v"
	mp4 = "mp4"
	MOV = "mov"
)
const VIDEO = "video"

func init() {
	Index.MergeCommands(ice.Commands{
		VIDEO: {Name: "video url run", Help: "视频", Actions: ice.MergeActions(ctx.ConfAction(nfs.PATH, ice.USR_LOCAL_IMAGE)), Hand: func(m *ice.Message, arg ...string) {
			_image_show(m, arg[0], arg[1:]...)
		}},
	})
}
