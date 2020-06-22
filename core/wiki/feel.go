package wiki

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"os"
	"path"
)

func _feel_show(m *ice.Message, name string, arg ...string) {
	m.Echo(path.Join(m.Conf(FEEL, "meta.path"), name))
}

const FEEL = "feel"
const (
	FeelPlugin = "/plugin/local/wiki/feel.js"
)

func init() {
	Index.Merge(&ice.Context{Name: "feel", Help: "影音媒体",
		Configs: map[string]*ice.Config{
			FEEL: {Name: "feel", Help: "影音媒体", Value: kit.Data(
				kit.MDB_SHORT, "name", "path", "", "regs", ".*.(qrc|png|jpg|JPG|MOV|m4v)",
			)},
		},
		Commands: map[string]*ice.Command{
			FEEL: {Name: "feel path=auto auto", Help: "影音媒体", Meta: kit.Dict(
				mdb.PLUGIN, FeelPlugin, "detail", []string{"标签", "删除"},
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
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if !_wiki_list(m, FEEL, kit.Select("./", arg, 0)) {
					_feel_show(m, arg[0])
				}
			}},
		},
	}, nil)
}
