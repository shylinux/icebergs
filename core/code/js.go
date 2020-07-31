package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"net/http"
	"os"
	"path"
	"strings"
)

func _js_find(m *ice.Message, key string) {
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, "find", ".", "-name", key), "\n") {
		if p == "" {
			continue
		}
		m.Push("file", strings.TrimPrefix(p, "./"))
		m.Push("line", 1)
		m.Push("text", "")
	}
}
func _js_grep(m *ice.Message, key string) {
	m.Split(m.Cmd(cli.SYSTEM, "grep", "--exclude-dir=.git", "--exclude=.[a-z]*", "-rn", key, ".").Append(cli.CMD_OUT), "file:line:text", ":", "\n")
}

const JS = "js"
const TS = "ts"
const TSX = "tsx"
const CSS = "css"
const HTML = "html"
const NODE = "node"

func init() {
	Index.Register(&ice.Context{Name: JS, Help: "js",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, JS, JS, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, JS, JS, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, JS, JS, c.Cap(ice.CTX_FOLLOW))
			}},
			NODE: {Name: NODE, Help: "node", Action: map[string]*ice.Action{
				"install": {Name: "install", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					// 下载
					source := m.Conf(NODE, "meta.source")
					p := path.Join(m.Conf("web.code._install", "meta.path"), path.Base(source))
					if _, e := os.Stat(p); e != nil {
						msg := m.Cmd(web.SPIDE, "dev", web.CACHE, http.MethodGet, source)
						m.Cmd(web.CACHE, web.WATCH, msg.Append(web.DATA), p)
					}

					// 解压
					m.Option(cli.CMD_DIR, m.Conf("web.code._install", "meta.path"))
					m.Cmd(cli.SYSTEM, "tar", "xvf", path.Base(source))
					m.Echo(p)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

			JS: {Name: JS, Help: "js", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_js_find(m, kit.Select("main", arg, 1))
					_js_grep(m, kit.Select("main", arg, 1))
				}},
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(JS, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		},
		Configs: map[string]*ice.Config{
			NODE: {Name: NODE, Help: "服务器", Value: kit.Data(
				"source", "https://nodejs.org/dist/v10.13.0/node-v10.13.0-linux-x64.tar.xz",
			)},
			JS: {Name: JS, Help: "js", Value: kit.Data(
				"plug", kit.Dict(
					"split", kit.Dict(
						"space", " \t",
						"operator", "{[(&.,;!|<>)]}",
					),
					"prefix", kit.Dict(
						"//", "comment",
						"/*", "comment",
						"*", "comment",
					),
					"keyword", kit.Dict(
						"var", "keyword",
						"new", "keyword",
						"delete", "keyword",
						"typeof", "keyword",
						"function", "keyword",

						"if", "keyword",
						"else", "keyword",
						"for", "keyword",
						"while", "keyword",
						"break", "keyword",
						"continue", "keyword",
						"switch", "keyword",
						"case", "keyword",
						"default", "keyword",
						"return", "keyword",

						"window", "function",
						"console", "function",
						"document", "function",
						"arguments", "function",
						"event", "function",
						"Date", "function",
						"JSON", "function",

						"0", "string",
						"1", "string",
						"10", "string",
						"-1", "string",
						"true", "string",
						"false", "string",
						"undefined", "string",
						"null", "string",

						"__proto__", "function",
						"setTimeout", "function",
						"createElement", "function",
						"appendChild", "function",
						"removeChild", "function",
						"parentNode", "function",
						"childNodes", "function",

						"Volcanos", "function",
						"request", "function",
						"require", "function",

						"cb", "function",
						"cbs", "function",
						"shy", "function",
						"can", "function",
						"sub", "function",
						"msg", "function",
						"res", "function",
						"pane", "function",
						"plugin", "function",

						"-1", "string",
						"0", "string",
						"1", "string",
						"2", "string",
					),
				),
			)},
		},
	}, nil)
}
