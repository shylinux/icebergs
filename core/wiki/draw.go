package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DRAW = "draw"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DRAW: {Name: DRAW, Help: "思维导图", Value: kit.Data(lex.REGEXP, ".*\\.svg")},
	}, Commands: map[string]*ice.Command{
		DRAW: {Name: "draw path=src/main.svg pid refresh:button=auto edit save actions", Help: "思维导图", Meta: kit.Dict(ice.DisplayLocal("")), Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, mdb.TYPE, "svg", mdb.NAME, m.PrefixKey())
			}},
			mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				m.Echo("<html><body>")
				defer m.Echo("</body></html>")
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
			nfs.SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				_wiki_save(m, DRAW, arg[0], m.Option("content"))
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if !_wiki_list(m, DRAW, kit.Select(nfs.PWD, arg, 0)) {
				_wiki_show(m, DRAW, arg[0])
			}
		}},
	}})
}
