package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const SHY = "shy"

func init() {
	Index.Register(&ice.Context{Name: SHY, Help: "脚本", Commands: map[string]*ice.Command{
		SHY: {Name: SHY, Help: "脚本", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{mdb.SEARCH, mdb.ENGINE, mdb.RENDER, mdb.PLUGIN} {
					m.Cmd(cmd, mdb.CREATE, SHY, m.PrefixKey())
				}
				LoadPlug(m, SHY)
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessCommand("web.wiki.word", kit.Simple(path.Join(arg[2], arg[1])))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.SOURCE, path.Join(arg[2], arg[1]), kit.Dict(ice.MSG_ALIAS, m.Confv("web.wiki.word", kit.Keym(mdb.ALIAS))))
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == SHY {
					_go_find(m, kit.Select(cli.MAIN, arg, 1), arg[2])
					_go_grep(m, kit.Select(cli.MAIN, arg, 1), arg[2])
				}
			}},
		}, PlugAction())},
	}, Configs: map[string]*ice.Config{
		SHY: {Name: SHY, Help: "脚本", Value: kit.Data(PLUG, kit.Dict(
			mdb.RENDER, kit.Dict(),
			PREFIX, kit.Dict("# ", COMMENT), PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"source", "return",
					"title", "premenu", "chapter", "section",
					"refer", "spark", "field",
					"chart", "label", "chain",
					"image",
				),
			), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
