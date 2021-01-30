package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

func _inner_ext(name string) string {
	return strings.ToLower(kit.Select(path.Base(name), strings.TrimPrefix(path.Ext(name), ".")))
}
func _inner_list(m *ice.Message, ext, file, dir string, arg ...string) {
	if m.Warn(!m.Right(dir, file), ice.ErrNotRight, path.Join(dir, file)) {
		return // 没有权限
	}
	if m.Cmdy(mdb.RENDER, ext, file, dir, arg); m.Result() != "" {
		return // 解析成功
	}

	if m.Conf(INNER, kit.Keys(kit.META_SOURCE, ext)) == "true" {
		m.Cmdy(nfs.CAT, path.Join(dir, file))
	}
}
func _inner_show(m *ice.Message, ext, file, dir string, arg ...string) {
	if m.Warn(!m.Right(dir, file), ice.ErrNotRight, path.Join(dir, file)) {
		return // 没有权限
	}
	if m.Cmdy(mdb.ENGINE, ext, file, dir, arg); m.Result() != "" {
		return // 执行成功
	}

	if ls := kit.Simple(m.Confv(INNER, kit.Keym("show", ext))); len(ls) > 0 {
		m.Option(cli.CMD_DIR, dir)
		m.Cmdy(cli.SYSTEM, ls, file)
		m.Set(ice.MSG_APPEND)
	}
}

const (
	COMMENT  = "comment"
	KEYWORD  = "keyword"
	FUNCTION = "function"
	DATATYPE = "datatype"
	STRING   = "string"
)
const (
	SPLIT  = "split"
	PREFIX = "prefix"
	SUFFIX = "suffix"
)

const INNER = "inner"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			INNER: {Name: "inner path=src/ file=main.go line=1 auto", Help: "源代码", Meta: kit.Dict(
				"display", "/plugin/local/code/inner.js", "style", "editor",
			), Action: map[string]*ice.Action{
				mdb.PLUGIN: {Name: "plugin", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					if m.Cmdy(mdb.PLUGIN, arg); m.Result() == "" {
						m.Echo(kit.Select("{}", m.Conf(INNER, kit.Keym("plug", arg[0]))))
					}
				}},
				mdb.ENGINE: {Name: "engine", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					_inner_show(m, arg[0], arg[1], arg[2])
				}},
				mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					if strings.Contains(arg[1], ";") {
						ls := strings.Split(arg[1], ";")
						arg[0] = ls[0]
						arg[1] = ls[1]
					}
					m.Option(cli.CMD_DIR, arg[2])
					m.Option(nfs.DIR_ROOT, arg[2])
					m.Cmdy(mdb.SEARCH, arg[:2], "cmd,file,line,text")
				}},
				mdb.INPUTS: {Name: "favor inputs", Help: "补全"},

				FAVOR:       {Name: "favor", Help: "收藏"},
				ctx.COMMAND: {Name: "command", Help: "命令"},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					m.Cmdy(nfs.DIR, kit.Select("./", arg, 0))
					return
				}
				_inner_list(m, _inner_ext(arg[1]), arg[1], arg[0])
			}},
		},
		Configs: map[string]*ice.Config{
			INNER: {Name: "inner", Help: "源代码", Value: kit.Data(
				"source", kit.Dict(
					"s", "true", "S", "true",
					"shy", "true", "py", "true",
					"csv", "true", "json", "true",
					"css", "true", "html", "true",
					"txt", "true", "url", "true",
					"log", "true", "err", "true",

					"md", "true", "license", "true", "makefile", "true",
					"ini", "true", "conf", "true", "toml", "true",
				),
				"plug", kit.Dict(
					"s", kit.Dict(
						PREFIX, kit.Dict("//", COMMENT),
						KEYWORD, kit.Dict(
							"TEXT", KEYWORD,
							"RET", KEYWORD,
						),
					),
					"S", kit.Dict(
						PREFIX, kit.Dict("//", COMMENT),
						KEYWORD, kit.Dict(),
					),
					"py", kit.Dict(
						PREFIX, kit.Dict("#", COMMENT),
						KEYWORD, kit.Dict("print", KEYWORD),
					),
					"html", kit.Dict(
						SPLIT, kit.Dict(
							"space", " ",
							"operator", "<>",
						),
						KEYWORD, kit.Dict(
							"head", KEYWORD,
							"body", KEYWORD,
						),
					),
					"css", kit.Dict(
						SUFFIX, kit.Dict("{", COMMENT),
					),

					"md", kit.Dict(),
					"makefile", kit.Dict(
						PREFIX, kit.Dict("#", COMMENT),
						SUFFIX, kit.Dict(":", COMMENT),
						KEYWORD, kit.Dict(
							"ifeq", KEYWORD,
							"ifneq", KEYWORD,
							"else", KEYWORD,
							"endif", KEYWORD,
						),
					),
				),
				"show", kit.Dict(
					"py", []string{"python"},
					"js", []string{"node"},
				),
			)},
		},
	})
}
