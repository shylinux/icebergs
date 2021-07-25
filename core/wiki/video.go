package wiki

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

var video = `<video class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
title="{{.Option "text"}}" src="{{.Option "text"}}" controls></video>`

func _video_show(m *ice.Message, name, text string, arg ...string) {
	if !strings.HasPrefix(text, "http") && !strings.HasPrefix(text, "/") {
		text = path.Join("/share/local", _wiki_path(m, FEEL, text))
	}

	_option(m, VIDEO, name, text, arg...)
	m.RenderTemplate(m.Conf(VIDEO, kit.Keym(kit.MDB_TEMPLATE)))
}

const (
	mp4 = "mp4"
	m4v = "m4v"
	MOV = "mov"
)

const VIDEO = "video"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			VIDEO: {Name: "video", Help: "视频", Value: kit.Data(kit.MDB_TEMPLATE, video)},
		},
		Commands: map[string]*ice.Command{
			VIDEO: {Name: "video [name] url", Help: "视频", Action: map[string]*ice.Action{
				mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_video_show(m, arg[1], path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_video_show(m, arg[0], arg[1], arg[2:]...)
			}},
		}})
}
