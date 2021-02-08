package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const FEEL = "feel"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FEEL: {Name: FEEL, Help: "影音媒体", Value: kit.Data(
				kit.MDB_PATH, "usr/image", "regs", ".*.(png|PNG|jpg|JPG|jpeg|mp4|m4v|MOV)",
			)},
		},
		Commands: map[string]*ice.Command{
			FEEL: {Name: "feel path auto upload 上一页 下一页 参数", Help: "影音媒体", Meta: kit.Dict(
				"display", "/plugin/local/wiki/feel.js",
			), Action: map[string]*ice.Action{
				web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					_wiki_upload(m, FEEL, m.Option(kit.MDB_PATH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if !_wiki_list(m, FEEL, kit.Select("./", arg, 0)) {
				}
			}},
		},
	})
}
