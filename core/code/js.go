package code

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const JS = "js"
const CSS = "css"
const HTML = "html"
const NODE = "node"
const VUE = "vue"

func init() {
	Index.Register(&ice.Context{Name: JS, Help: "前端",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.PLUGIN, mdb.CREATE, JS, m.Prefix(JS))
				m.Cmd(mdb.RENDER, mdb.CREATE, JS, m.Prefix(JS))
				m.Cmd(mdb.ENGINE, mdb.CREATE, JS, m.Prefix(JS))
				m.Cmd(mdb.SEARCH, mdb.CREATE, JS, m.Prefix(JS))

				m.Cmd(mdb.PLUGIN, mdb.CREATE, VUE, m.Prefix(VUE))
				m.Cmd(mdb.RENDER, mdb.CREATE, VUE, m.Prefix(VUE))
				m.Cmd(mdb.ENGINE, mdb.CREATE, VUE, m.Prefix(VUE))
				m.Cmd(mdb.SEARCH, mdb.CREATE, VUE, m.Prefix(VUE))
			}},
			JS: {Name: JS, Help: "前端", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(JS, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
				mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, arg[2])
					m.Cmdy(cli.SYSTEM, NODE, arg[1])
					m.Set(ice.MSG_APPEND)
				}},
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					_go_find(m, kit.Select("main", arg, 1))
					_go_grep(m, kit.Select("main", arg, 1))
				}},
			}},
			NODE: {Name: "node", Help: "前端", Action: map[string]*ice.Action{
				web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(INSTALL, m.Conf(NODE, kit.META_SOURCE))
				}},
			}},
			VUE: {Name: "vue", Help: "前端", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(VUE, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
				mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, arg[2])
					m.Cmdy(cli.SYSTEM, NODE, arg[1])
					m.Set(ice.MSG_APPEND)
				}},
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					_go_find(m, kit.Select("main", arg, 1))
					_go_grep(m, kit.Select("main", arg, 1))
				}},
			}},
		},
		Configs: map[string]*ice.Config{
			VUE: {Name: VUE, Help: "vue", Value: kit.Data(
				"plug", kit.Dict(
					PREFIX, kit.Dict(
						"//", COMMENT,
						"/*", COMMENT,
						"*", COMMENT,
					),
					SPLIT, kit.Dict(
						"space", " \t",
						"operator", "{[(&.,;!|<>)]}",
					),
				),
			)},
			NODE: {Name: NODE, Help: "前端", Value: kit.Data(
				kit.SSH_SOURCE, "https://nodejs.org/dist/v10.13.0/node-v10.13.0-linux-x64.tar.xz",
			)},
			JS: {Name: JS, Help: "js", Value: kit.Data(
				"plug", kit.Dict(
					SPLIT, kit.Dict(
						"space", " \t",
						"operator", "{[(&.,;!|<>)]}",
					),
					PREFIX, kit.Dict(
						"//", COMMENT,
						"/*", COMMENT,
						"*", COMMENT,
					),
					KEYWORD, kit.Dict(
						"import", KEYWORD,
						"from", KEYWORD,
						"export", KEYWORD,

						"var", KEYWORD,
						"new", KEYWORD,
						"delete", KEYWORD,
						"typeof", KEYWORD,
						"const", KEYWORD,
						"function", KEYWORD,

						"if", KEYWORD,
						"else", KEYWORD,
						"for", KEYWORD,
						"while", KEYWORD,
						"break", KEYWORD,
						"continue", KEYWORD,
						"switch", KEYWORD,
						"case", KEYWORD,
						"default", KEYWORD,
						"return", KEYWORD,
						"try", KEYWORD,
						"throw", KEYWORD,
						"catch", KEYWORD,
						"finally", KEYWORD,

						"window", FUNCTION,
						"console", FUNCTION,
						"document", FUNCTION,
						"arguments", FUNCTION,
						"event", FUNCTION,
						"Date", FUNCTION,
						"JSON", FUNCTION,

						"0", STRING,
						"1", STRING,
						"10", STRING,
						"-1", STRING,
						"true", STRING,
						"false", STRING,
						"undefined", STRING,
						"null", STRING,

						"__proto__", FUNCTION,
						"setTimeout", FUNCTION,
						"createElement", FUNCTION,
						"appendChild", FUNCTION,
						"removeChild", FUNCTION,
						"parentNode", FUNCTION,
						"childNodes", FUNCTION,

						"Volcanos", FUNCTION,
						"request", FUNCTION,
						"require", FUNCTION,

						"cb", FUNCTION,
						"cbs", FUNCTION,
						"shy", FUNCTION,
						"can", FUNCTION,
						"sub", FUNCTION,
						"msg", FUNCTION,
						"res", FUNCTION,
						"pane", FUNCTION,
						"plugin", FUNCTION,

						"-1", STRING,
						"0", STRING,
						"1", STRING,
						"2", STRING,
					),
				),
			)},
		},
	}, nil)
}
