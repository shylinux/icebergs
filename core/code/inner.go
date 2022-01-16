package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	kit "shylinux.com/x/toolkits"
)

func _inner_exec(m *ice.Message, ext, file, dir string, arg ...string) {
	if !m.Right(dir, file) {
		return // 没有权限
	}
	if m.Cmdy(mdb.ENGINE, ext, file, dir, arg); m.Result() != "" {
		return // 执行成功
	}
	if ls := kit.Simple(m.Confv(INNER, kit.Keym(EXEC, ext))); len(ls) > 0 {
		m.Cmdy(cli.SYSTEM, ls, file, ice.Option{cli.CMD_DIR, dir})
		m.Set(ice.MSG_APPEND)
	}
}
func _inner_list(m *ice.Message, ext, file, dir string, arg ...string) {
	if !m.Right(dir, file) {
		return // 没有权限
	}
	if m.Cmdy(mdb.RENDER, ext, file, dir, arg); m.Result() != "" {
		return // 解析成功
	}
	if m.Config(kit.Keys(ssh.SOURCE, ext)) == ice.TRUE {
		m.Cmdy(nfs.CAT, path.Join(dir, file))
	}
}

func LoadPlug(m *ice.Message, language string) {
	m.Confm(language, kit.Keym(PLUG, PREPARE), func(key string, value interface{}) {
		for _, v := range kit.Simple(value) {
			m.Conf(language, kit.Keym(PLUG, KEYWORD, v), key)
		}
	})
}

func PlugAction(fields ...string) map[string]*ice.Action {
	return ice.SelectAction(map[string]*ice.Action{
		mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Config(PLUG)) }},
		mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
	}, fields...)
}

const (
	COMMENT  = "comment"
	KEYWORD  = "keyword"
	DATATYPE = "datatype"
	FUNCTION = "function"
	CONSTANT = "constant"
)
const (
	SPLIT  = "split"
	PREFIX = "prefix"
	SUFFIX = "suffix"
)
const (
	PLUG = "plug"
	EXEC = "exec"
)
const INNER = "inner"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		INNER: {Name: "inner path=src/@key file=main.go line=1 auto", Help: "源代码", Meta: kit.Dict(
			ice.DisplayLocal(""),
		), Action: ice.MergeAction(map[string]*ice.Action{
			mdb.PLUGIN: {Name: "plugin", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(mdb.PLUGIN, arg); m.Result() == "" {
					m.Echo(kit.Select("{}", m.Config(kit.Keys(PLUG, arg[0]))))
				}
				m.Set(ice.MSG_STATUS)
			}},
			mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, kit.Path(arg[2]))
				m.Option(nfs.DIR_ROOT, arg[2])
				m.Cmdy(mdb.SEARCH, arg[0], arg[1], arg[2])
			}},
			mdb.ENGINE: {Name: "engine", Help: "引擎", Hand: func(m *ice.Message, arg ...string) {
				_inner_exec(m, arg[0], arg[1], arg[2])
			}},
			mdb.INPUTS: {Name: "favor inputs", Help: "补全"},
			FAVOR:      {Name: "favor", Help: "收藏"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if !strings.HasSuffix(arg[0], ice.PS) {
				arg[1] = kit.Slice(strings.Split(arg[0], ice.PS), -1)[0]
				arg[0] = strings.TrimSuffix(arg[0], arg[1])
				m.ProcessRewrite(nfs.PATH, arg[0], nfs.FILE, arg[1])
				return
			}
			if len(arg) < 2 {
				nfs.Dir(m, nfs.PATH)
				m.Set(ice.MSG_STATUS)
				return
			}
			m.Option("plug", "inner/search.js?a=1,inner/favor.js,inner/sess.js")
			arg[1] = kit.Split(arg[1])[0]
			_inner_list(m, kit.Ext(arg[1]), arg[1], arg[0])
			m.Set(ice.MSG_STATUS)
		}},
	}, Configs: map[string]*ice.Config{
		INNER: {Name: "inner", Help: "源代码", Value: kit.Data(
			ssh.SOURCE, kit.Dict(
				"s", ice.TRUE, "S", ice.TRUE,
				"shy", ice.TRUE, "py", ice.TRUE,
				"csv", ice.TRUE, "json", ice.TRUE,
				"css", ice.TRUE, "html", ice.TRUE,
				"txt", ice.TRUE, "url", ice.TRUE,
				"log", ice.TRUE, "err", ice.TRUE,

				"md", ice.TRUE, "license", ice.TRUE, "makefile", ice.TRUE, "sql", ice.TRUE,
				"ini", ice.TRUE, "conf", ice.TRUE, "toml", ice.TRUE, "yaml", ice.TRUE, "yml", ice.TRUE,
			),
			EXEC, kit.Dict(
				"py", []string{"python"},
				"js", []string{"node"},
			),
			PLUG, kit.Dict(
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
				"yaml", kit.Dict(
					PREFIX, kit.Dict("#", COMMENT),
				),
				"yml", kit.Dict(
					PREFIX, kit.Dict("#", COMMENT),
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
		)},
	}})
}
