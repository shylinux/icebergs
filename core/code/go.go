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
	fields := kit.Split(m.Option(mdb.FIELDS))
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, "find", ".", "-name", key), "\n") {
		if p == "" {
			continue
		}
		for _, k := range fields {
			switch k {
			case kit.MDB_FILE:
				m.Push(k, strings.TrimPrefix(p, "./"))
			case kit.MDB_LINE:
				m.Push(k, 1)
			case kit.MDB_TEXT:
				m.Push(k, "")
			default:
				m.Push(k, "")
			}
		}
	}
}
func _go_tags(m *ice.Message, key string) {
	if _, e := os.Stat(path.Join(m.Option(cli.CMD_DIR), ".tags")); e != nil {
		m.Cmd(cli.SYSTEM, "gotags", "-R", "-f", ".tags", "./")
	}

	fields := kit.Split(m.Option(mdb.FIELDS))
	ls := strings.Split(key, ".")
	key = ls[len(ls)-1]

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

		f, e := os.Open(path.Join(m.Option(cli.CMD_DIR), file))
		m.Assert(e)
		defer f.Close()

		bio := bufio.NewScanner(f)
		for i := 1; bio.Scan(); i++ {
			if i == line || bio.Text() == text {
				for _, k := range fields {
					switch k {
					case kit.MDB_FILE:
						m.Push(k, strings.TrimPrefix(file, "./"))
					case kit.MDB_LINE:
						m.Push(k, i)
					case kit.MDB_TEXT:
						m.Push(k, bio.Text())
					default:
						m.Push(k, "")
					}
				}
			}
		}
	}
}
func _go_grep(m *ice.Message, key string) {
	fields := kit.Split(m.Option(mdb.FIELDS))

	msg := m.Spawn()
	msg.Split(m.Cmd(cli.SYSTEM, "grep", "--exclude-dir=.git", "--exclude=.[a-z]*", "-rn", key, ".").Append(cli.CMD_OUT), "file:line:text", ":", "\n")
	msg.Table(func(index int, value map[string]string, head []string) {
		m.Push("", value, fields)
	})
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

	for _, k := range kit.Split(m.Option(mdb.FIELDS)) {
		switch k {
		case kit.MDB_FILE:
			m.Push(k, key+".godoc")
		case kit.MDB_LINE:
			m.Push(k, 1)
		case kit.MDB_TEXT:
			m.Push(k, string(res))
		default:
			m.Push(k, "")
		}
	}
}

const GO = "go"
const DOC = "godoc"
const MOD = "mod"
const SUM = "sum"

func init() {
	Index.Register(&ice.Context{Name: GO, Help: "后端",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.PLUGIN, mdb.CREATE, GO, m.Prefix(GO))
				m.Cmd(mdb.RENDER, mdb.CREATE, GO, m.Prefix(GO))
				m.Cmd(mdb.ENGINE, mdb.CREATE, GO, m.Prefix(GO))
				m.Cmd(mdb.SEARCH, mdb.CREATE, GO, m.Prefix(GO))

				m.Cmd(mdb.PLUGIN, mdb.CREATE, DOC, m.Prefix(DOC))
				m.Cmd(mdb.RENDER, mdb.CREATE, DOC, m.Prefix(DOC))
				m.Cmd(mdb.SEARCH, mdb.CREATE, DOC, m.Prefix(GO))

				m.Cmd(mdb.PLUGIN, mdb.CREATE, MOD, m.Prefix(MOD))
				m.Cmd(mdb.RENDER, mdb.CREATE, MOD, m.Prefix(MOD))

				m.Cmd(mdb.PLUGIN, mdb.CREATE, SUM, m.Prefix(SUM))
				m.Cmd(mdb.RENDER, mdb.CREATE, SUM, m.Prefix(SUM))

			}},
			SUM: {Name: SUM, Help: "版本", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(MOD, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}},
			MOD: {Name: MOD, Help: "模块", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(MOD, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}},
			DOC: {Name: DOC, Help: "文档", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(GO, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, arg[2])
					m.Echo(m.Cmdx(cli.SYSTEM, GO, "doc", strings.TrimSuffix(arg[1], "."+arg[0])))
				}},
			}},
			GO: {Name: GO, Help: "后端", Action: map[string]*ice.Action{
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
					m.Set(ice.MSG_APPEND)
				}},
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					m.Option(cli.CMD_DIR, kit.Select("src", arg, 2))
					_go_find(m, kit.Select("main", arg, 1))
					_go_help(m, kit.Select("main", arg, 1))
					_go_tags(m, kit.Select("main", arg, 1))
					_go_grep(m, kit.Select("main", arg, 1))
				}},
			}},
		},
		Configs: map[string]*ice.Config{
			MOD: {Name: MOD, Help: "模块", Value: kit.Data(
				"plug", kit.Dict(
					PREFIX, kit.Dict(
						"#", COMMENT,
					),
					KEYWORD, kit.Dict(
						"module", KEYWORD,
						"require", KEYWORD,
						"replace", KEYWORD,
						"=>", KEYWORD,
					),
				),
			)},
			GO: {Name: GO, Help: "后端", Value: kit.Data(
				"plug", kit.Dict(
					SPLIT, kit.Dict(
						"space", "\t ",
						"operator", "{[(&.,:;!|<>)]}",
					),
					PREFIX, kit.Dict(
						"//", COMMENT,
						"/*", COMMENT,
						"*", COMMENT,
					),
					KEYWORD, kit.Dict(
						"package", KEYWORD,
						"import", KEYWORD,
						"const", KEYWORD,
						"func", KEYWORD,
						"var", KEYWORD,
						"type", KEYWORD,
						"struct", KEYWORD,
						"interface", KEYWORD,

						"if", KEYWORD,
						"else", KEYWORD,
						"for", KEYWORD,
						"range", KEYWORD,
						"break", KEYWORD,
						"continue", KEYWORD,
						"switch", KEYWORD,
						"case", KEYWORD,
						"default", KEYWORD,
						"fallthrough", KEYWORD,

						"go", KEYWORD,
						"select", KEYWORD,
						"return", KEYWORD,
						"defer", KEYWORD,

						"map", DATATYPE,
						"chan", DATATYPE,
						"string", DATATYPE,
						"error", DATATYPE,
						"bool", DATATYPE,
						"byte", DATATYPE,
						"int", DATATYPE,
						"int64", DATATYPE,
						"float64", DATATYPE,

						"len", FUNCTION,
						"cap", FUNCTION,
						"copy", FUNCTION,
						"append", FUNCTION,
						"msg", FUNCTION,
						"m", FUNCTION,

						"nil", STRING,
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
