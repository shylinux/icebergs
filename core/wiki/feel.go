package wiki

import (
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const FEEL = "feel"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FEEL: {Name: FEEL, Help: "影音媒体", Value: kit.Data(nfs.PATH, ice.USR_LOCAL_IMAGE, lex.REGEXP, ".*.(png|PNG|jpg|JPG|jpeg|mp4|m4v|MOV)")},
	}, Commands: map[string]*ice.Command{
		FEEL: {Name: "feel path auto upload 上一页 下一页 actions", Help: "影音媒体", Meta: kit.Dict(ice.DisplayLocal("")), Action: map[string]*ice.Action{
			web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				_wiki_upload(m, m.CommandKey(), m.Option(nfs.PATH))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				p := path.Join(m.Config(nfs.PATH), path.Join(arg...))
				m.Debug("remove %v", p)
				os.Remove(p)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_wiki_list(m, m.CommandKey(), kit.Select(nfs.PWD, arg, 0))
		}},
	}})
}
