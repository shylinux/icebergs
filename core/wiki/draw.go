package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DRAW = "draw"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DRAW: {Name: DRAW, Help: "思维导图", Value: kit.Data(REGEXP, ".*\\.svg")},
	}, Commands: map[string]*ice.Command{
		DRAW: {Name: "draw path=src/main.svg pid refresh:button=auto edit save actions", Help: "思维导图", Meta: kit.Dict(
			ice.DisplayLocal(""),
		), Action: ice.MergeAction(map[string]*ice.Action{
			nfs.SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				_wiki_save(m, DRAW, arg[0], m.Option(kit.MDB_CONTENT))
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if !_wiki_list(m, DRAW, kit.Select(ice.PWD, arg, 0)) {
				_wiki_show(m, DRAW, arg[0])
			}
		}},
	}})
}
