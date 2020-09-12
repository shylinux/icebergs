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

func _inner_ext(name string) string {
	return strings.ToLower(kit.Select(path.Base(name), strings.TrimPrefix(path.Ext(name), ".")))
}
func _inner_list(m *ice.Message, ext, file, dir string, arg ...string) {
	if m.Warn(!m.Right(dir, file), ice.ErrNotAuth, path.Join(dir, file)) {
		return
	}
	if m.Cmdy(mdb.RENDER, ext, file, dir, arg); m.Result() != "" {
		return
	}
	if m.Conf(INNER, kit.Keys("meta.source", ext)) == "true" {
		if m.Cmdy(mdb.RENDER, nfs.FILE, file, dir, arg); m.Result() != "" {
			return
		}
	}
	m.Echo(path.Join(dir, file))
}
func _inner_show(m *ice.Message, ext, file, dir string, arg ...string) {
	if m.Cmdy(mdb.ENGINE, ext, file, dir, arg); m.Result() == "" {
		if ls := kit.Simple(m.Confv(INNER, kit.Keys("meta.show", ext))); len(ls) > 0 {
			m.Cmdy(cli.SYSTEM, ls, path.Join(dir, file)).Set(ice.MSG_APPEND)
		}
	}
}
func _vimer_save(m *ice.Message, ext, file, dir string, text string) {
	if f, p, e := kit.Create(path.Join(dir, file)); e == nil {
		defer f.Close()
		if n, e := f.WriteString(text); m.Assert(e) {
			m.Log_EXPORT("file", path.Join(dir, file), "size", n)
		}
		m.Echo(p)
	}
}

const INNER = "inner"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			INNER: {Name: "inner path=usr/demo file=hi.sh line=1 auto 运行:button 项目:button 搜索:button", Help: "阅读器", Meta: kit.Dict(
				"display", "/plugin/local/code/inner.js", "style", "editor",
			), Action: map[string]*ice.Action{
				mdb.PLUGIN: {Name: "plugin type name text arg...", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					if m.Cmdy(mdb.PLUGIN, arg); m.Result() == "" {
						if m.Echo(m.Conf(INNER, kit.Keys("meta.plug", arg[0]))); m.Result() == "" {
							m.Echo("{}")
						}
					}
				}},
				mdb.SEARCH: {Name: "search type name text arg...", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.SEARCH, arg)
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
				if strings.HasPrefix("http", arg[0]) {
					m.Cmdy(web.SPIDE, "dev", "raw", "GET", arg[0]+arg[1])
					return
				}
				_inner_list(m, _inner_ext(arg[1]), arg[1], arg[0])
			}},
		},
		Configs: map[string]*ice.Config{
			INNER: {Name: "inner", Help: "阅读器", Value: kit.Data(
				"source", kit.Dict(
					"license", "true",
					"makefile", "true",
					"shy", "true", "py", "true",
					"csv", "true", "json", "true",
					"css", "true", "html", "true",
					"txt", "true", "url", "true",
					"log", "true", "err", "true",

					"md", "true", "conf", "true", "toml", "true",
					"ts", "true", "tsx", "true", "vue", "true", "sass", "true",
				),
				"plug", kit.Dict(
					"makefile", kit.Dict(
						"prefix", kit.Dict("#", "comment"),
						"suffix", kit.Dict(":", "comment"),
						"keyword", kit.Dict(
							"ifeq", "keyword",
							"ifneq", "keyword",
							"else", "keyword",
							"endif", "keyword",
						),
					),
					"py", kit.Dict(
						"prefix", kit.Dict("#", "comment"),
						"keyword", kit.Dict("print", "keyword"),
					),
					"csv", kit.Dict("display", true),
					"json", kit.Dict("link", true),
					"html", kit.Dict(
						"split", kit.Dict(
							"space", " ",
							"operator", "<>",
						),
						"keyword", kit.Dict(
							"head", "keyword",
							"body", "keyword",
						),
					),
					"css", kit.Dict(
						"suffix", kit.Dict("{", "comment"),
					),

					"md", kit.Dict("display", true, "profile", true),
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
