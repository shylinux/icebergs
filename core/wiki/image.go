package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _image_show(m *ice.Message, text string) {
	_wiki_template(m, IMAGE, "", _wiki_link(m, IMAGE, text))
}

const (
	IMG  = "img"
	PNG  = "png"
	JPG  = "jpg"
	JPEG = "jpeg"
)
const IMAGE = "image"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		IMAGE: {Name: "image url height auto", Help: "图片", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, PNG, m.PrefixKey())
			}},
			mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_image_show(m, path.Join(arg[2], arg[1]))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				return
			}
			m.Option("height", kit.Select("", arg, 1))
			_image_show(m, arg[0])
		}},
	}, Configs: map[string]*ice.Config{
		IMAGE: {Name: IMAGE, Help: "图片", Value: kit.Data(
			nfs.TEMPLATE, `<img {{.OptionTemplate}} title="{{.Option "text"}}" src="{{.Option "text"}}" height="{{.Option "height"}}">`,
			nfs.PATH, ice.USR_LOCAL_IMAGE,
		)},
	}})
}
