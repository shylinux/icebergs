package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _shy_exec(m *ice.Message, arg ...string) {
	switch left := kit.Select("", kit.Slice(kit.Split(m.Option(mdb.TEXT), "\t \n`"), -1), 0); strings.TrimSpace(left) {
	case cli.FG, cli.BG:
		m.Push(mdb.NAME, cli.RED)
		m.Push(mdb.NAME, cli.BLUE)
		m.Push(mdb.NAME, cli.GREEN)

	default:
		switch kit.Select("", kit.Split(m.Option(mdb.TEXT)), 0) {
		case "field":
			m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "", ice.OptionFields("index,name,text"))
			_vimer_list(m, ice.SRC, ctx.INDEX)

		case "chain":
			m.Push(mdb.NAME, cli.FG)
			m.Push(mdb.NAME, cli.BG)
		}
	}
}

const SHY = "shy"

func init() {
	Index.Register(&ice.Context{Name: SHY, Help: "脚本", Commands: ice.Commands{
		SHY: {Name: "shy path auto", Help: "脚本", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{mdb.SEARCH, mdb.ENGINE, mdb.RENDER, mdb.PLUGIN} {
					m.Cmd(cmd, mdb.CREATE, SHY, m.PrefixKey())
				}
				LoadPlug(m, SHY)
				gdb.Watch(m, VIMER_TEMPLATE)
			}},
			VIMER_TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				if kit.Ext(m.Option(mdb.FILE)) != m.CommandKey() {
					return
				}
				m.Echo(`
chapter "hi"
`)
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessCommand(m, "web.wiki.word", kit.Simple(path.Join(arg[2], arg[1])))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				_shy_exec(m, arg...)
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == SHY {
					_go_find(m, kit.Select(cli.MAIN, arg, 1), arg[2])
					_go_grep(m, kit.Select(cli.MAIN, arg, 1), arg[2])
				}
			}},
		}, PlugAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == m.CommandKey() {
				m.Cmdy("web.wiki.word", path.Join(ice.SRC, arg[0]))
				return
			}
			m.Cmdy("web.wiki.word", arg)
		}},
	}, Configs: ice.Configs{
		SHY: {Name: SHY, Help: "脚本", Value: kit.Data(PLUG, kit.Dict(
			mdb.RENDER, kit.Dict(),
			PREFIX, kit.Dict("# ", COMMENT), PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"source", "return",
					"title", "premenu", "chapter", "section",
					"refer", "spark", "field",
					"chart", "label", "chain", "sequence",
					"image",
				),
			), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
