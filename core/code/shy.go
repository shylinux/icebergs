package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"path"
)

const SHY = "shy"

func init() {
	Index.Register(&ice.Context{Name: SHY, Help: "shy",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, SHY, SHY, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, SHY, SHY, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, SHY, SHY, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.ENGINE, mdb.CREATE, SHY, SHY, c.Cap(ice.CTX_FOLLOW))
			}},
			SHY: {Name: SHY, Help: "shy", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_c_find(m, kit.Select("main", arg, 1))
					_c_grep(m, kit.Select("main", arg, 1))
				}},
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(SHY, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
				mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.wiki.word", path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		},
		Configs: map[string]*ice.Config{
			SHY: {Name: SHY, Help: "shy", Value: kit.Data(
				"plug", kit.Dict(
					"prefix", kit.Dict("#", "comment"),
					"keyword", kit.Dict(
						"title", "keyword",
						"chapter", "keyword",
						"section", "keyword",
						"refer", "keyword",
						"field", "keyword",
						"label", "keyword",
						"chain", "keyword",
					),
				),
			)},
		},
	}, nil)
}
