package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

const (
	INNER  = "inner"
	VEDIO  = "vedio"
	QRCODE = "qrcode"
)

const (
	LIST = "list"
	PLUG = "plug"
	SHOW = "show"
	SAVE = "save"
)

func _inner_ext(name string) string {
	return strings.ToLower(kit.Select(path.Base(name), strings.TrimPrefix(path.Ext(name), ".")))
}

func _inner_show(m *ice.Message, ext, file, dir string, arg ...string) {
	if m.Cmdy(mdb.RENDER, ext, file, dir, arg); m.Result() == "" {
		if ls := kit.Simple(m.Confv(INNER, kit.Keys("meta.show", ext))); len(ls) > 0 {
			m.Cmdy(cli.SYSTEM, ls, path.Join(dir, file)).Set(ice.MSG_APPEND)
		}
	}
}
func _inner_list(m *ice.Message, ext, file, dir string, arg ...string) {
	if m.Cmdy(mdb.RENDER, ext, file, dir, arg); m.Result() == "" {
		if m.Conf(INNER, kit.Keys("meta.source", ext)) == "true" {
			if m.Cmdy(mdb.RENDER, nfs.FILE, file, dir, arg); m.Result() == "" {
				m.Echo(path.Join(dir, file))
			}
		}
	}
}

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			INNER: {Name: "inner path=usr/demo file=hi.qrc line=1 查看:button=auto", Help: "编辑器", Meta: kit.Dict(
				"display", "/plugin/local/code/inner.js", "style", "editor",
			), Action: map[string]*ice.Action{
				web.UPLOAD: {Name: "upload path name", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.CACHE, web.UPLOAD)
					m.Cmdy(web.CACHE, web.WATCH, m.Option(web.DATA), path.Join(m.Option("path"), m.Option("name")))
				}},

				mdb.SEARCH: {Name: "search type name text arg...", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.SEARCH, arg)
				}},
				mdb.PLUGIN: {Name: "plugin type name text arg...", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					if m.Cmdy(mdb.PLUGIN, arg); m.Result() == "" {
						if m.Echo(m.Conf(INNER, kit.Keys("meta.plug", arg[0]))); m.Result() == "" {
							m.Echo("{}")
						}
					}
				}},
				mdb.RENDER: {Name: "render type name text arg...", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_inner_list(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},
				mdb.ENGINE: {Name: "engine type name text arg...", Help: "引擎", Hand: func(m *ice.Message, arg ...string) {
					_inner_show(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					_inner_list(m, nfs.DIR, "", kit.Select("", arg, 0))
					return
				}
				_inner_list(m, _inner_ext(arg[1]), arg[1], arg[0])
			}},
		},
		Configs: map[string]*ice.Config{
			INNER: {Name: "inner", Help: "编辑器", Value: kit.Data(
				"source", kit.Dict(
					"makefile", "true",
					"shy", "true", "py", "true",

					"md", "true", "csv", "true",
					"txt", "true", "url", "true",
					"conf", "true", "json", "true",
					"ts", "true", "tsx", "true", "vue", "true", "sass", "true",
					"html", "true", "css", "true",
				),
				"plug", kit.Dict(
					"py", kit.Dict(
						"prefix", kit.Dict("#", "comment"),
						"keyword", kit.Dict("print", "keyword"),
					),
					"md", kit.Dict("display", true, "profile", true),
					"csv", kit.Dict("display", true),
					"ts", kit.Dict(
						"prefix", kit.Dict("//", "comment"),
						"split", kit.Dict(
							"space", " ",
							"operator", "{[(.:,;!|)]}",
						),
						"keyword", kit.Dict(
							"import", "keyword",
							"from", "keyword",
							"new", "keyword",
							"as", "keyword",
							"const", "keyword",
							"export", "keyword",
							"default", "keyword",

							"if", "keyword",
							"return", "keyword",

							"class", "keyword",
							"extends", "keyword",
							"interface", "keyword",
							"declare", "keyword",
							"async", "keyword",
							"await", "keyword",
							"try", "keyword",
							"catch", "keyword",

							"function", "function",
							"arguments", "function",
							"console", "function",
							"this", "function",

							"string", "datatype",
							"number", "datatype",

							"true", "string",
							"false", "string",
						),
					),
					"tsx", kit.Dict("link", "ts"),
					"vue", kit.Dict("link", "ts"),
					"sass", kit.Dict("link", "ts"),
				),
				"show", kit.Dict(
					"sh", []string{"sh"},
					"py", []string{"python"},
					"js", []string{"node"},
				),
			)},
		},
	}, nil)
}
