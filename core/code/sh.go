package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _sh_main_script(m *ice.Message, arg ...string) (res []string) {
	if kit.FileExists(kit.Path(arg[2], arg[1])) {
		res = append(res, kit.Format("source %s", kit.Path(arg[2], arg[1])))
	} else if b, ok := ice.Info.Pack[path.Join(arg[2], arg[1])]; ok && len(b) > 0 {
		res = append(res, string(b))
	}
	return
}

const SH = "sh"

func init() {
	Index.Register(&ice.Context{Name: SH, Help: "命令", Commands: map[string]*ice.Command{
		SH: {Name: SH, Help: "命令", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{mdb.SEARCH, mdb.ENGINE, mdb.RENDER, mdb.PLUGIN} {
					m.Cmd(cmd, mdb.CREATE, SH, m.PrefixKey())
				}
				LoadPlug(m, SH)
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, SH, "-c", kit.Join(_sh_main_script(m, arg...), ice.NL)).SetAppend()
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, SH, "-c", kit.Join(_sh_main_script(m, arg...), ice.NL)).SetAppend()
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH {
					return
				}
				m.Option(cli.CMD_DIR, kit.Select(ice.SRC, arg, 2))
				m.Cmdy(mdb.SEARCH, MAN1, arg[1:])
				m.Cmdy(mdb.SEARCH, MAN8, arg[1:])
				_go_find(m, kit.Select(cli.MAIN, arg, 1), arg[2])
				_go_grep(m, kit.Select(cli.MAIN, arg, 1), arg[2])
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
