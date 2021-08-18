package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const SHY = "shy"

func init() {
	Index.Register(&ice.Context{Name: SHY, Help: "脚本",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				for _, cmd := range []string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE, mdb.SEARCH} {
					m.Cmd(cmd, mdb.CREATE, SHY, m.Prefix(SHY))
				}
				LoadPlug(m, SHY)
			}},
			SHY: {Name: SHY, Help: "脚本", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(SHY, kit.Keym(PLUG)))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
				mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.wiki.word", path.Join(arg[2], arg[1]))
				}},
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					_go_find(m, kit.Select(kit.MDB_MAIN, arg, 1))
					_go_grep(m, kit.Select(kit.MDB_MAIN, arg, 1))
				}},
			}},
		},
		Configs: map[string]*ice.Config{
			SHY: {Name: SHY, Help: "脚本", Value: kit.Data(
				PLUG, kit.Dict(
					PREFIX, kit.Dict("#", COMMENT),
					PREPARE, kit.Dict(
						KEYWORD, kit.Simple(
							"title",
							"premenu",
							"chapter",
							"section",
							"source",
							"refer",
							"field",
							"spark",
							"image",
							"label",
							"chain",
						),
					),
					KEYWORD, kit.Dict(),
				),
			)},
		},
	}, nil)
}
