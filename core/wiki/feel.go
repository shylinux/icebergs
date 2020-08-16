package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
)

const FEEL = "feel"

func init() {
	Index.Merge(&ice.Context{Name: "feel", Help: "影音媒体",
		Configs: map[string]*ice.Config{
			FEEL: {Name: "feel", Help: "影音媒体", Value: kit.Data(
				"path", "usr/image", "regs", ".*.(qrc|png|PNG|jpg|jpeg|JPG|MOV|m4v|mp4)",
			)},
		},
		Commands: map[string]*ice.Command{
			FEEL: {Name: "feel path=auto auto 上传:button 上一页:button 下一页:button 参数:button", Help: "影音媒体", Meta: kit.Dict(
				"display", "/plugin/local/wiki/feel.js", "detail", []string{"标签", "删除"},
			), Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Conf(FEEL, kit.Keys(path.Base(arg[2]), "-2"), arg[3])
					p := path.Join(m.Conf(FEEL, "meta.path"), arg[2])
					q := path.Join(m.Conf(FEEL, "meta.path"), arg[3])
					os.MkdirAll(q, 0777)
					m.Assert(os.Link(p, path.Join(q, path.Base(arg[2]))))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Assert(os.Remove(path.Join(m.Conf(FEEL, "meta.path"), m.Option("path"))))
				}},
				web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					_wiki_upload(m, FEEL)
				}},
				web.SPIDE: {Name: "spide type title url poster", Help: "爬虫", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.SPIDE, "self", "cache", "GET", arg[2])
					m.Cmd(web.CACHE, "watch", m.Append("data"), path.Join(m.Conf(FEEL, "meta.path"), m.Option("path"), arg[1])+path.Ext(arg[2]))
					if m.Option("path") != "最近" {
						m.Cmd(web.CACHE, "watch", m.Append("data"), path.Join(m.Conf(FEEL, "meta.path"), "最近", arg[1])+path.Ext(arg[2]))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option("prefix", m.Conf(FEEL, "meta.path"))
				if !_wiki_list(m, FEEL, kit.Select("./", arg, 0)) {
					m.Echo(path.Join(m.Conf(FEEL, "meta.path"), arg[0]))
					return
				}
				m.Sort("time", "time_r")
			}},
		},
	}, nil)
}
