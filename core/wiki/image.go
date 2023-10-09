package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

func _image_show(m *ice.Message, name, text string, arg ...string) {
	_text := text
	nfs.Exists(m, path.Join(ice.USR_LOCAL_IMAGE, text), func(p string) { text = p })
	nfs.Exists(m, path.Join(ice.USR_ICONS, text), func(p string) { text = p })
	_wiki_template(m.Options(web.LINK, _wiki_link(m, text)), "", name, _text, arg...)
}

const IMAGE = "image"

func init() {
	Index.MergeCommands(ice.Commands{
		IMAGE: {Name: "image path", Help: "图片", Hand: func(m *ice.Message, arg ...string) {
			arg = _name(m, arg)
			_image_show(m, arg[0], arg[1], arg[2:]...)
		}},
	})
}
