package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
)

const (
	mp4 = "mp4"
	m4v = "m4v"
	MOV = "mov"
)
const VIDEO = "video"

func init() {
	Index.MergeCommands(ice.Commands{
		VIDEO: {Name: "video url", Help: "视频", Actions: WordAction(
			`<video {{.OptionTemplate}} title="{{.Option "text"}}" src="{{.Option "text"}}" controls></video>`, nfs.PATH, ice.USR_LOCAL_IMAGE,
		), Hand: func(m *ice.Message, arg ...string) { _image_show(m, arg[0], arg[1:]...) }},
	})
}
