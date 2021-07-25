package code

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

const SH = "sh"

func init() {
	Index.Register(&ice.Context{Name: SH, Help: "命令",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				for _, cmd := range []string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE, mdb.SEARCH} {
					m.Cmd(cmd, mdb.CREATE, SH, m.Prefix(SH))
				}
				LoadPlug(m, SH)
			}},
			SH: {Name: SH, Help: "命令", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(SH, kit.Keym(PLUG)))
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
					_go_find(m, kit.Select(kit.MDB_MAIN, arg, 1))
					m.Cmdy(mdb.SEARCH, MAN1, arg[1:])
					m.Cmdy(mdb.SEARCH, MAN8, arg[1:])
					_go_grep(m, kit.Select(kit.MDB_MAIN, arg, 1))
				}},

				MAN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(_c_help(m, arg[0], arg[1]))
				}},
			}},
		},
		Configs: map[string]*ice.Config{
			SH: {Name: SH, Help: "命令", Value: kit.Data(
				PLUG, kit.Dict(
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
					PREPARE, kit.Dict(
						KEYWORD, kit.Simple(
							"export",
							"source",
							"require",

							"if",
							"then",
							"else",
							"fi",
							"for",
							"while",
							"do",
							"done",
							"esac",
							"case",
							"in",
							"return",

							"shift",
							"local",
							"echo",
							"eval",
							"kill",
							"let",
							"cd",
						),
						FUNCTION, kit.Simple(
							"xargs",
							"date",
							"find",
							"grep",
							"sed",
							"awk",
							"pwd",
							"ps",
							"ls",
							"rm",
							"go",
						),
					),
					KEYWORD, kit.Dict(),
				),
			)},
		},
	}, nil)
}
