package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DRAW = "draw"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DRAW: {Name: DRAW, Help: "思维导图", Value: kit.Data(kit.MDB_REGEXP, ".*\\.svg")},
	}, Commands: map[string]*ice.Command{
		DRAW: {Name: "draw path=src/main.svg refresh:button=auto edit save project actions", Help: "思维导图", Meta: kit.Dict(
			ice.Display("/plugin/local/wiki/draw.js"),
		), Action: ice.MergeAction(map[string]*ice.Action{
			nfs.SAVE: {Name: "save path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				_wiki_save(m, DRAW, path.Join(arg...), m.Option(kit.MDB_CONTENT))
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if !_wiki_list(m, DRAW, kit.Select("./", path.Join(arg...))) {
				_wiki_show(m, DRAW, path.Join(arg...))
			}
		}},
	}})
}
