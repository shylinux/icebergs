package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/format"
)

func init() {
	format.RegisterAll()

	Index.Register(&ice.Context{Name: "m4v", Help: "视频",
		Configs: map[string]*ice.Config{
			VIDEO: {Name: "video", Help: "视频", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			"list": {Name: "list name", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(arg[0])
			}},
			"save": {Name: "save name text", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy("nfs.qrcodes", arg)
			}},
			"show": {Name: "show name", Help: "渲染", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if file, e := avutil.Open(arg[0]); m.Assert(e) {
					defer file.Close()
					if streams, e := file.Streams(); m.Assert(e) {
						for _, stream := range streams {
							if stream.Type().IsAudio() {

							} else if stream.Type().IsVideo() {
								vstream := stream.(av.VideoCodecData)
								m.Push("type", vstream.Type().String())
								m.Push("width", vstream.Width())
								m.Push("height", vstream.Height())
							}
						}
					}
				}
			}},
			"video": {Name: "video", Help: "视频", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(web.FAVOR, arg, "extra", "extra.poster").Table(func(index int, value map[string]string, header []string) {
					m.Echo(`<video src="%s" controls loop></video>`, value["text"])
				})
			}},
		},
	}, nil)
}