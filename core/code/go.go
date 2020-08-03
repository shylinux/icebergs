package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"os"
	"path"
	"strings"
)

func _go_find(m *ice.Message, key string) {
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, "find", ".", "-name", key), "\n") {
		if p == "" {
			continue
		}
		m.Push("file", strings.TrimPrefix(p, "./"))
		m.Push("line", 1)
		m.Push("text", "")
	}
}
func _go_tags(m *ice.Message, key string) {
	ls := strings.Split(key, ".")
	key = ls[len(ls)-1]

	if _, e := os.Stat(path.Join(m.Option("_path"), ".tags")); e != nil {
		m.Cmd(cli.SYSTEM, "gotags", "-R", "-f", ".tags", "./")
	}
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, "grep", "^"+key+"\\>", ".tags"), "\n") {
		ls := strings.SplitN(l, "\t", 2)
		if len(ls) < 2 {
			continue
		}

		ls = strings.SplitN(ls[1], "\t", 2)
		file := ls[0]
		ls = strings.SplitN(ls[1], ";\"", 2)
		text := strings.TrimSuffix(strings.TrimPrefix(ls[0], "/^"), "$/")
		line := kit.Int(text)

		p := path.Join(m.Option("_path"), file)
		f, e := os.Open(p)
		m.Assert(e)
		bio := bufio.NewScanner(f)
		for i := 1; bio.Scan(); i++ {
			if i == line || bio.Text() == text {
				m.Push("file", strings.TrimPrefix(file, "./"))
				m.Push("line", i)
				m.Push("text", bio.Text())
			}
		}
	}
	m.Sort("line", "int")
}
func _go_grep(m *ice.Message, key string) {
	m.Split(m.Cmd(cli.SYSTEM, "grep", "--exclude-dir=.git", "--exclude=.[a-z]*", "-rn", key, ".").Append(cli.CMD_OUT), "file:line:text", ":", "\n")
}
func _go_help(m *ice.Message, key string) {
	p := m.Cmd(cli.SYSTEM, "go", "doc", key).Append(cli.CMD_OUT)
	if p == "" {
		return
	}
	ls := strings.Split(p, "\n")
	if len(ls) > 10 {
		ls = ls[:10]
	}
	res := strings.Join(ls, "\n")

	m.Push("file", key+".godoc")
	m.Push("line", 1)
	m.Push("text", string(res))
}

const GO = "go"
const GODOC = "godoc"
const MOD = "mod"
const SUM = "sum"

func init() {
	Index.Register(&ice.Context{Name: GO, Help: "go",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, GO, GO, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, GO, GO, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, GO, GO, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.ENGINE, mdb.CREATE, GO, GO, c.Cap(ice.CTX_FOLLOW))

				m.Cmd(mdb.SEARCH, mdb.CREATE, GODOC, GO, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, GODOC, GO, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, GODOC, GODOC, c.Cap(ice.CTX_FOLLOW))

				m.Cmd(mdb.PLUGIN, mdb.CREATE, MOD, MOD, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, MOD, MOD, c.Cap(ice.CTX_FOLLOW))

				m.Cmd(mdb.PLUGIN, mdb.CREATE, SUM, SUM, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, SUM, SUM, c.Cap(ice.CTX_FOLLOW))

			}},
			MOD: {Name: MOD, Help: "mod", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(GO, "meta.mod.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}},
			SUM: {Name: SUM, Help: "sum", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(GO, "meta.mod.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}},
			GODOC: {Name: GODOC, Help: "godoc", Action: map[string]*ice.Action{
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, arg[2])
					m.Echo(m.Cmdx(cli.SYSTEM, GO, "doc", strings.TrimSuffix(arg[1], "."+arg[0])))
				}},
			}},
			GO: {Name: GO, Help: "go", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_go_find(m, kit.Select("main", arg, 1))
					_go_tags(m, kit.Select("main", arg, 1))
					_go_help(m, kit.Select("main", arg, 1))
					_go_grep(m, kit.Select("main", arg, 1))
				}},
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(GO, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
				mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, arg[2])
					if strings.HasSuffix(arg[1], "test.go") {
						m.Cmdy(cli.SYSTEM, GO, "test", "-v", "./"+arg[1])
					} else {
						m.Cmdy(cli.SYSTEM, GO, "run", "./"+arg[1])
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		},
		Configs: map[string]*ice.Config{
			GO: {Name: GO, Help: "go", Value: kit.Data(
				"mod.plug", kit.Dict(
					"prefix", kit.Dict(
						"#", "comment",
					),
					"keyword", kit.Dict(
						"module", "keyword",
						"require", "keyword",
						"replace", "keyword",
						"=>", "keyword",
					),
				),
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
						"package", "keyword",
						"import", "keyword",
						"const", "keyword",
						"func", "keyword",
						"var", "keyword",
						"type", "keyword",
						"struct", "keyword",
						"interface", "keyword",

						"if", "keyword",
						"else", "keyword",
						"for", "keyword",
						"range", "keyword",
						"break", "keyword",
						"continue", "keyword",
						"switch", "keyword",
						"case", "keyword",
						"default", "keyword",
						"fallthrough", "keyword",

						"go", "keyword",
						"select", "keyword",
						"return", "keyword",
						"defer", "keyword",

						"map", "datatype",
						"chan", "datatype",
						"string", "datatype",
						"error", "datatype",
						"bool", "datatype",
						"byte", "datatype",
						"int", "datatype",
						"int64", "datatype",
						"float64", "datatype",

						"len", "function",
						"cap", "function",
						"copy", "function",
						"append", "function",
						"msg", "function",
						"m", "function",

						"nil", "string",
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
