package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
)

const AUDIO = "audio"

func init() {
	Index.MergeCommands(ice.Commands{
		AUDIO: {Name: "audio url run", Help: "音频", Actions: WordAction(
			`<audio {{.OptionTemplate}} title="{{.Option "text"}}" src="{{.Option "text"}}" controls></audio>`, nfs.PATH, ice.USR_LOCAL_IMAGE,
		), Hand: func(m *ice.Message, arg ...string) { _image_show(m, arg[0], arg[1:]...) }},
	})
}
