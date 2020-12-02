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
	Index.Register(&ice.Context{Name: SHY, Help: "脚本",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.PLUGIN, mdb.CREATE, SHY, SHY, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, SHY, SHY, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.ENGINE, mdb.CREATE, SHY, SHY, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, SHY, SHY, c.Cap(ice.CTX_FOLLOW))
			}},
			SHY: {Name: SHY, Help: "脚本", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(SHY, "meta.plug"))
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
					m.Option(cli.CMD_DIR, kit.Select("src", arg, 2))
					_go_find(m, kit.Select("main", arg, 1))
					_go_grep(m, kit.Select("main", arg, 1))
				}},
			}},
		},
		Configs: map[string]*ice.Config{
			SHY: {Name: SHY, Help: "脚本", Value: kit.Data(
				"plug", kit.Dict(
					PREFIX, kit.Dict("#", COMMENT),
					KEYWORD, kit.Dict(
						"title", KEYWORD,
						"premenu", KEYWORD,
						"chapter", KEYWORD,
						"section", KEYWORD,
						"source", KEYWORD,
						"refer", KEYWORD,
						"field", KEYWORD,
						"label", KEYWORD,
						"chain", KEYWORD,
					),
				),
			)},
		},
	}, nil)
}
