package wiki

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _image_show(m *ice.Message, text string, arg ...string) {
	_wiki_template(m, IMAGE, "", _wiki_link(m, IMAGE, text), arg...)
}

const (
	PNG  = "png"
	JPG  = "jpg"
	JPEG = "jpeg"
)
const IMAGE = "image"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, PNG, m.Prefix(IMAGE))
			}},
			IMAGE: {Name: "image url", Help: "图片", Action: map[string]*ice.Action{
				mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_image_show(m, path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_image_show(m, arg[0], arg[1:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			IMAGE: {Name: IMAGE, Help: "图片", Value: kit.Data(
				kit.MDB_TEMPLATE, `<img {{.OptionTemplate}} title="{{.Option "text"}}" src="{{.Option "text"}}">`,
				kit.MDB_PATH, "usr/local/image",
			)},
		},
	})
}
