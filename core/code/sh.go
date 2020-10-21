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
	Index.Register(&ice.Context{Name: SH, Help: "命令",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.PLUGIN, mdb.CREATE, SH, SH, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, SH, SH, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.ENGINE, mdb.CREATE, SH, SH, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, SH, SH, c.Cap(ice.CTX_FOLLOW))
			}},
			SH: {Name: SH, Help: "命令", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(SH, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
				mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, arg[2])
					m.Cmdy(cli.SYSTEM, SH, arg[1])
					m.Set(ice.MSG_APPEND)
				}},
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					m.Option(cli.CMD_DIR, kit.Select("src", arg, 2))
					_go_find(m, kit.Select("main", arg, 1))
					m.Cmdy(mdb.SEARCH, MAN1, arg[1:])
					m.Cmdy(mdb.SEARCH, MAN8, arg[1:])
					_go_grep(m, kit.Select("main", arg, 1))
				}},

				MAN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(_c_help(m, arg[0], arg[1]))
				}},
			}},
		},
		Configs: map[string]*ice.Config{
			SH: {Name: SH, Help: "命令", Value: kit.Data(
				"plug", kit.Dict(
					SPLIT, kit.Dict(
						"space", " ",
						"operator", "{[(.,;!|<>)]}",
					),
					PREFIX, kit.Dict(
						"#", COMMENT,
					),
					SUFFIX, kit.Dict(
						"{", COMMENT,
					),
					KEYWORD, kit.Dict(
						"export", KEYWORD,
						"source", KEYWORD,
						"require", KEYWORD,

						"if", KEYWORD,
						"then", KEYWORD,
						"else", KEYWORD,
						"fi", KEYWORD,
						"for", KEYWORD,
						"while", KEYWORD,
						"do", KEYWORD,
						"done", KEYWORD,
						"esac", KEYWORD,
						"case", KEYWORD,
						"in", KEYWORD,
						"return", KEYWORD,

						"shift", KEYWORD,
						"local", KEYWORD,
						"echo", KEYWORD,
						"eval", KEYWORD,
						"kill", KEYWORD,
						"let", KEYWORD,
						"cd", KEYWORD,

						"xargs", FUNCTION,
						"date", FUNCTION,
						"find", FUNCTION,
						"grep", FUNCTION,
						"sed", FUNCTION,
						"awk", FUNCTION,
						"pwd", FUNCTION,
						"ps", FUNCTION,
						"ls", FUNCTION,
						"rm", FUNCTION,
						"go", FUNCTION,
					),
				),
			)},
		},
	}, nil)
}
