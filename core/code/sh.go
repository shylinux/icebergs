package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"path"
)

const SH = "sh"

func init() {
	Index.Register(&ice.Context{Name: SH, Help: "sh",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, SH, SH, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, SH, SH, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, SH, SH, c.Cap(ice.CTX_FOLLOW))
			}},
			SH: {Name: SH, Help: "sh", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_c_find(m, kit.Select("main", arg, 1))
					m.Cmdy(mdb.SEARCH, "man1", arg[1:])
					_c_grep(m, kit.Select("main", arg, 1))
				}},
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(SH, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
				mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		},
		Configs: map[string]*ice.Config{
			SH: {Name: SH, Help: "sh", Value: kit.Data(
				"plug", kit.Dict(
					"split", kit.Dict(
						"space", " ",
						"operator", "{[(.,;!|<>)]}",
					),
					"prefix", kit.Dict(
						"#", "comment",
					),
					"suffix", kit.Dict(
						"{", "comment",
					),
					"keyword", kit.Dict(
						"export", "keyword",
						"source", "keyword",
						"require", "keyword",

						"if", "keyword",
						"then", "keyword",
						"else", "keyword",
						"fi", "keyword",
						"for", "keyword",
						"while", "keyword",
						"do", "keyword",
						"done", "keyword",
						"esac", "keyword",
						"case", "keyword",
						"in", "keyword",
						"return", "keyword",

						"shift", "keyword",
						"local", "keyword",
						"echo", "keyword",
						"eval", "keyword",
						"kill", "keyword",
						"let", "keyword",
						"cd", "keyword",

						"xargs", "function",
						"date", "function",
						"find", "function",
						"grep", "function",
						"sed", "function",
						"awk", "function",
						"pwd", "function",
						"ps", "function",
						"ls", "function",
						"rm", "function",
						"go", "function",
					),
				),
			)},
		},
	}, nil)
}
