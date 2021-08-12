package code

import (
	"bufio"
	"os"
	"path"
	"strings"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

func _go_find(m *ice.Message, key string) {
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, "find", ".", "-name", key), "\n") {
		if p == "" {
			continue
		}
		m.PushSearch(cli.CMD, "find", kit.MDB_FILE, strings.TrimPrefix(p, "./"), kit.MDB_LINE, 1, kit.MDB_TEXT, "")
	}
}
func _go_tags(m *ice.Message, key string) {
	if s, e := os.Stat(path.Join(m.Option(cli.CMD_DIR), ".tags")); os.IsNotExist(e) || s.ModTime().Before(time.Now().Add(kit.Duration("-72h"))) {
		m.Cmd(cli.SYSTEM, "gotags", "-R", "-f", ".tags", "./")
	}

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
				m.PushSearch(cli.CMD, "tags", kit.MDB_FILE, strings.TrimPrefix(file, "./"), kit.MDB_LINE, kit.Format(i), kit.MDB_TEXT, bio.Text())
			}
		}
	}
}
func _go_grep(m *ice.Message, key string) {
	msg := m.Spawn()
	msg.Split(m.Cmd(cli.SYSTEM, "grep", "--exclude-dir=.git", "--exclude=.[a-z]*", "-rn", key, ".").Append(cli.CMD_OUT), "file:line:text", ":", "\n")
	msg.Table(func(index int, value map[string]string, head []string) { m.PushSearch(cli.CMD, "grep", value) })
}
func _go_help(m *ice.Message, key string) {
	p := m.Cmd(cli.SYSTEM, "go", "doc", key).Append(cli.CMD_OUT)
	if p == "" {
		return
	}
	m.PushSearch(cli.CMD, "help", kit.MDB_FILE, key+".godoc", kit.MDB_LINE, 1, kit.MDB_TEXT, p)
}

const GO = "go"
const DOC = "godoc"
const MOD = "mod"
const SUM = "sum"
const PROTO = "proto"

func init() {
	Index.Register(&ice.Context{Name: GO, Help: "后端",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.ENGINE, mdb.CREATE, GO, m.Prefix(GO))
				m.Cmd(mdb.SEARCH, mdb.CREATE, GO, m.Prefix(GO))
				m.Cmd(mdb.SEARCH, mdb.CREATE, DOC, m.Prefix(GO))

				for k := range c.Configs {
					m.Cmd(mdb.PLUGIN, mdb.CREATE, k, m.Prefix(k))
					m.Cmd(mdb.RENDER, mdb.CREATE, k, m.Prefix(k))
				}
				LoadPlug(m, GO)
			}},
			SUM: {Name: SUM, Help: "版本", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(MOD, kit.Keym(PLUG)))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}},
			MOD: {Name: MOD, Help: "模块", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(MOD, kit.Keym(PLUG)))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}},
			PROTO: {Name: PROTO, Help: "协议", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(PROTO, kit.Keym(PLUG)))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}},
			DOC: {Name: DOC, Help: "文档", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(GO, kit.Keym(PLUG)))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, arg[2])
					m.Echo(m.Cmdx(cli.SYSTEM, GO, "doc", strings.TrimSuffix(arg[1], "."+arg[0])))
				}},
			}},
			GO: {Name: GO, Help: "后端", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(GO, kit.Keym(PLUG)))
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
					_go_find(m, kit.Select(kit.MDB_MAIN, arg, 1))
					_go_help(m, kit.Select(kit.MDB_MAIN, arg, 1))
					_go_tags(m, kit.Select(kit.MDB_MAIN, arg, 1))
					_go_grep(m, kit.Select(kit.MDB_MAIN, arg, 1))
				}},
			}},
		},
		Configs: map[string]*ice.Config{
			PROTO: {Name: PROTO, Help: "协议", Value: kit.Data(
				PLUG, kit.Dict(
					PREFIX, kit.Dict(
						"//", COMMENT,
					),
					PREPARE, kit.Dict(
						KEYWORD, kit.Simple(
							"syntax", "option", "package", "import", "service", "message",
						),
						DATATYPE, kit.Simple(
							"string", "int64", "int32",
						),
					),
					KEYWORD, kit.Dict(),
				),
			)},
			MOD: {Name: MOD, Help: "模块", Value: kit.Data(
				PLUG, kit.Dict(
					PREFIX, kit.Dict(
						"//", COMMENT,
					),
					PREPARE, kit.Dict(
						KEYWORD, kit.Simple(
							"module", "require", "replace", "=>",
						),
					),
					KEYWORD, kit.Dict(),
				),
			)},
			GO: {Name: GO, Help: "后端", Value: kit.Data(
				PLUG, kit.Dict(
					SPLIT, kit.Dict(
						"space", "\t ",
						"operator", "{[(&.,:;!|<>)]}",
					),
					PREFIX, kit.Dict(
						"//", COMMENT,
						"/*", COMMENT,
						"* ", COMMENT,
					),
					PREPARE, kit.Dict(
						KEYWORD, kit.Simple(
							"package", "import", "type", "struct", "interface", "const", "var", "func",
							"if", "else", "for", "range", "break", "continue",
							"switch", "case", "default", "fallthrough",
							"go", "select", "defer", "return",
						),
						DATATYPE, kit.Simple(
							"int", "int32", "int64", "float64",
							"string", "byte", "bool", "error", "chan", "map",
						),
						FUNCTION, kit.Simple(
							"new", "make", "len", "cap", "copy", "append", "delete", "msg", "m",
						),
						CONSTANT, kit.Simple(
							"false", "true", "nil", "-1", "0", "1", "2",
						),
					),
					KEYWORD, kit.Dict(),
				),
			)},
		},
	}, nil)
}
