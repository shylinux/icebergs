package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const JS = "js"
const CSS = "css"
const HTML = "html"
const NODE = "node"
const VUE = "vue"

func init() {
	Index.Register(&ice.Context{Name: JS, Help: "前端", Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for _, cmd := range []string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE, mdb.SEARCH} {
				m.Cmd(cmd, mdb.CREATE, VUE, m.Prefix(JS))
				m.Cmd(cmd, mdb.CREATE, JS, m.Prefix(JS))
			}
			LoadPlug(m, JS)
		}},
		JS: {Name: JS, Help: "前端", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, arg[2])
				m.Cmdy(cli.SYSTEM, NODE, arg[1])
				m.Set(ice.MSG_APPEND)
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == kit.MDB_FOREACH {
					return
				}
				_go_find(m, kit.Select(kit.MDB_MAIN, arg, 1))
				_go_grep(m, kit.Select(kit.MDB_MAIN, arg, 1))
			}},
		}, PlugAction())},
		NODE: {Name: "node auto download", Help: "前端", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(INSTALL, m.Config(cli.SOURCE))
			}},
		}},
	}, Configs: map[string]*ice.Config{
		NODE: {Name: NODE, Help: "前端", Value: kit.Data(
			cli.SOURCE, "https://nodejs.org/dist/v10.13.0/node-v10.13.0-linux-x64.tar.xz",
		)},
		JS: {Name: JS, Help: "js", Value: kit.Data(PLUG, kit.Dict(
			SPLIT, kit.Dict("space", " \t", "operator", "{[(&.,;!|<>)]}"),
			PREFIX, kit.Dict("//", COMMENT, "/*", COMMENT, "*", COMMENT),
			PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"import",
					"from",
					"export",

					"var",
					"new",
					"delete",
					"typeof",
					"const",
					"function",

					"if",
					"else",
					"for",
					"while",
					"break",
					"continue",
					"switch",
					"case",
					"default",
					"return",
					"try",
					"throw",
					"catch",
					"finally",

					"can",
					"sub",
					"msg",
					"res",
					"target",
				),
				FUNCTION, kit.Simple(
					"window",
					"console",
					"document",
					"arguments",
					"event",
					"Date",
					"JSON",

					"__proto__",
					"setTimeout",
					"createElement",
					"appendChild",
					"removeChild",
					"parentNode",
					"childNodes",

					"Volcanos",
					"request",
					"require",

					"cb",
					"cbs",
					"shy",
					"pane",
					"plugin",
				),
				CONSTANT, kit.Simple(
					"true", "false",
					"undefined", "null",
					"-1", "0", "1", "2", "10",
				),
			), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
