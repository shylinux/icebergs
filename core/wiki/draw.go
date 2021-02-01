package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"path"
)

const DRAW = "draw"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DRAW: {Name: DRAW, Help: "思维导图", Value: kit.Data(kit.MDB_PATH, "", "regs", ".*\\.svg")},
		},
		Commands: map[string]*ice.Command{
			DRAW: {Name: "draw path=src/ file=main.svg 刷新:button=auto 编辑 save 项目 参数", Help: "思维导图", Meta: kit.Dict(
				"display", "/plugin/local/wiki/draw.js", "style", "drawer",
			), Action: map[string]*ice.Action{
				nfs.SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_wiki_save(m, DRAW, path.Join(arg...), m.Option("content"))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if !_wiki_list(m, DRAW, kit.Select("./", path.Join(arg...))) {
					_wiki_show(m, DRAW, path.Join(arg...))
				}
			}},
		},
	})
}
