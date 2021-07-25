package wiki

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _image_show(m *ice.Message, name, text string, arg ...string) {
	if !strings.HasPrefix(text, "http") && !strings.HasPrefix(text, "/") {
		text = path.Join("/share/local", _wiki_path(m, FEEL, text))
	}

	_option(m, IMAGE, name, text, arg...)
	m.RenderTemplate(m.Conf(IMAGE, kit.Keym(kit.MDB_TEMPLATE)))
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
			IMAGE: {Name: "image [name] url", Help: "图片", Action: map[string]*ice.Action{
				mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_image_show(m, arg[1], path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_image_show(m, arg[0], arg[1], arg[2:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			IMAGE: {Name: IMAGE, Help: "图片", Value: kit.Data(
				kit.MDB_TEMPLATE, `<img class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
title="{{.Option "text"}}" src="{{.Option "text"}}">`,
			)},
		},
	})
}
