package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TS = "ts"
const JS = "js"
const CSS = "css"
const HTML = "html"
const JSON = "json"
const NODE = "node"
const VUE = "vue"

func init() {
	Index.Register(&ice.Context{Name: JS, Help: "前端", Commands: map[string]*ice.Command{
		JS: {Name: "js", Help: "前端", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{mdb.SEARCH, mdb.ENGINE, mdb.RENDER, mdb.PLUGIN} {
					m.Cmd(cmd, mdb.CREATE, JSON, m.PrefixKey())
					m.Cmd(cmd, mdb.CREATE, VUE, m.PrefixKey())
					m.Cmd(cmd, mdb.CREATE, JS, m.PrefixKey())
					m.Cmd(cmd, mdb.CREATE, TS, m.PrefixKey())
				}
				LoadPlug(m, JS)
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH {
					return
				}
				_go_find(m, kit.Select(cli.MAIN, arg, 1), arg[2])
				_go_grep(m, kit.Select(cli.MAIN, arg, 1), arg[2])
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, NODE, arg[1], kit.Dict(cli.CMD_DIR, arg[2])).SetAppend()
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				if key, ok := ice.Info.File[kit.Replace(path.Join(arg[2], arg[1]), ".js", ".go")]; ok && key != "" {
					m.Display(path.Join(web.SHARE_LOCAL, path.Join(arg[2], arg[1])))
					m.ProcessCommand(key, kit.Simple())
				} else {
					m.Display(path.Join(web.SHARE_LOCAL, path.Join(arg[2], arg[1])))
					m.ProcessCommand("can.code.inner.plugin", kit.Simple())
					// m.ProcessCommand("web.wiki.word", kit.Simple(strings.ReplaceAll(path.Join(arg[2], arg[1]), ".go", ".shy")))
				}
			}},
		}, PlugAction())},
		NODE: {Name: "node auto download", Help: "前端", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(INSTALL, m.Config(nfs.SOURCE))
			}},
		}},
	}, Configs: map[string]*ice.Config{
		NODE: {Name: NODE, Help: "前端", Value: kit.Data(
			nfs.SOURCE, "https://nodejs.org/dist/v10.13.0/node-v10.13.0-linux-x64.tar.xz",
		)},
		JS: {Name: JS, Help: "js", Value: kit.Data(PLUG, kit.Dict(
			SPLIT, kit.Dict("space", " \t", "operator", "{[(&.,;!|<>)]}"),
			PREFIX, kit.Dict("//", COMMENT, "/*", COMMENT, "*", COMMENT), PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"import", "from", "export",

					"var", "new", "delete", "typeof", "const", "function",

					"if", "else", "for", "while", "break", "continue", "switch", "case", "default",
					"return", "try", "throw", "catch", "finally",

					"can", "sub", "msg", "res", "target",
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
