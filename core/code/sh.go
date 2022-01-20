package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const SH = "sh"

func init() {
	Index.Register(&ice.Context{Name: SH, Help: "命令", Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for _, cmd := range []string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE, mdb.SEARCH} {
				m.Cmd(cmd, mdb.CREATE, SH, m.Prefix(SH))
			}
			LoadPlug(m, SH)
		}},
		SH: {Name: SH, Help: "命令", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, arg[2])
				m.Cmdy(cli.SYSTEM, SH, arg[1])
				m.Set(ice.MSG_APPEND)
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH {
					return
				}
				m.Option(cli.CMD_DIR, kit.Select(ice.SRC, arg, 2))
				m.Cmdy(mdb.SEARCH, MAN1, arg[1:])
				m.Cmdy(mdb.SEARCH, MAN8, arg[1:])
				_go_find(m, kit.Select(MAIN, arg, 1), arg[2])
				_go_grep(m, kit.Select(MAIN, arg, 1), arg[2])
			}},
			MAN: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_c_help(m, arg[0], arg[1]))
			}},
		}, PlugAction())},
	}, Configs: map[string]*ice.Config{
		SH: {Name: SH, Help: "命令", Value: kit.Data(PLUG, kit.Dict(
			SPLIT, kit.Dict("space", " ", "operator", "{[(.,;!|<>)]}"),
			PREFIX, kit.Dict("#", COMMENT),
			SUFFIX, kit.Dict("{", COMMENT),
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
			), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
