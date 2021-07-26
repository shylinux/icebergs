package wiki

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _video_show(m *ice.Message, text string, arg ...string) {
	_wiki_template(m, VIDEO, "", _wiki_link(m, VIDEO, text), arg...)
}

const (
	mp4 = "mp4"
	m4v = "m4v"
	MOV = "mov"
)

const VIDEO = "video"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			VIDEO: {Name: "video url", Help: "视频", Action: map[string]*ice.Action{
				mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_video_show(m, path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_video_show(m, arg[0], arg[1:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			VIDEO: {Name: "video", Help: "视频", Value: kit.Data(
				kit.MDB_TEMPLATE, `<video {{.OptionTemplate}} title="{{.Option "text"}}" src="{{.Option "text"}}" controls></video>`,
				kit.MDB_PATH, "usr/local/image",
			)},
		},
	})
}
