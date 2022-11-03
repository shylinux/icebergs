package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
)

func _image_show(m *ice.Message, text string, arg ...string) {
	_wiki_template(m, "", _wiki_link(m, text), arg...)
}

const (
	IMG  = "img"
	PNG  = "png"
	JPG  = "jpg"
	JPEG = "jpeg"
)
const IMAGE = "image"

func init() {
	Index.MergeCommands(ice.Commands{
		IMAGE: {Name: "image url", Help: "图片", Actions: WordAction(
			`<img {{.OptionTemplate}} title="{{.Option "text"}}" src="{{.Option "text"}}">`, nfs.PATH, ice.USR_LOCAL_IMAGE,
		), Hand: func(m *ice.Message, arg ...string) { _image_show(m, arg[0], arg[1:]...) }},
	})
}
