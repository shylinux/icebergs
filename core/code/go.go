package code

import (
	"bufio"
	"os"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _go_tags(m *ice.Message, key string) {
	if s, e := os.Stat(path.Join(m.Option(cli.CMD_DIR), TAGS)); os.IsNotExist(e) || s.ModTime().Before(time.Now().Add(kit.Duration("-72h"))) {
		m.Cmd(cli.SYSTEM, "gotags", "-R", "-f", TAGS, nfs.PWD)
	}

	ls := strings.Split(key, ice.PT)
	key = ls[len(ls)-1]

	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, nfs.GREP, "^"+key+"\\>", TAGS), ice.NL) {
		ls := strings.SplitN(l, ice.TB, 2)
		if len(ls) < 2 {
			continue
		}

		ls = strings.SplitN(ls[1], ice.TB, 2)
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
				m.PushSearch(nfs.FILE, strings.TrimPrefix(file, nfs.PWD), nfs.LINE, kit.Format(i), mdb.TEXT, bio.Text())
			}
		}
	}
}
func _go_help(m *ice.Message, key string) {
	if p := m.Cmd(cli.SYSTEM, GO, "doc", key).Append(cli.CMD_OUT); strings.TrimSpace(p) != "" {
		m.PushSearch(nfs.FILE, key+".godoc", nfs.LINE, 1, mdb.TEXT, p)
	}
}
func _go_find(m *ice.Message, key string, dir string) {
	m.Cmd(nfs.FIND, dir, key).Table(func(index int, value map[string]string, head []string) {
		m.PushSearch(nfs.LINE, 1, value)
	})
}
func _go_grep(m *ice.Message, key string, dir string) {
	m.Cmd(nfs.GREP, dir, key).Table(func(index int, value map[string]string, head []string) {
		m.PushSearch(value)
	})
}
func _go_exec(m *ice.Message, arg ...string) {
	if key, ok := ice.Info.File[path.Join(arg[2], arg[1])]; ok && key != "" {
		m.Cmdy(cli.SYSTEM, GO, ice.RUN, ice.SRC_MAIN_GO, key)
	} else if m.Option(cli.CMD_DIR, arg[2]); strings.HasSuffix(arg[1], "_test.go") {
		m.Cmdy(cli.SYSTEM, GO, "test", "-v", nfs.PWD+arg[1])
	} else {
		m.Cmdy(cli.SYSTEM, GO, ice.RUN, nfs.PWD+arg[1])
	}
	m.SetAppend()
}
func _go_show(m *ice.Message, arg ...string) {
	if arg[1] == "main.go" {
		const (
			PACKAGE = "package"
			IMPORT  = "import"
		)
		block := ""
		index := 0
		push := func(repos string) {
			index++
			m.Push("index", index)
			m.Push("repos", repos)
		}
		m.Cmd(nfs.CAT, path.Join(arg[2], arg[1]), func(line string) {
			ls := kit.Split(line)
			switch {
			case strings.HasPrefix(line, IMPORT+" ("):
				block = IMPORT
				return
			case strings.HasPrefix(line, ")"):
				block = ""
				return
			case strings.HasPrefix(line, IMPORT):
				if len(ls) == 2 {
					push(ls[1])
				} else if len(ls) == 3 {
					push(ls[2])
				}
				return
			}
			switch block {
			case IMPORT:
				if len(ls) == 0 {
					push("")
				} else if len(ls) == 1 {
					push(ls[0])
				} else if len(ls) == 2 {
					push(ls[1])
				}
			}
		})
		return
	}
	if key, ok := ice.Info.File[path.Join(arg[2], arg[1])]; ok && key != "" {
		m.ProcessCommand(key, kit.Simple())
	} else {
		m.ProcessCommand("web.wiki.word", kit.Simple(strings.ReplaceAll(path.Join(arg[2], arg[1]), ".go", ".shy")))
	}
}
func _mod_show(m *ice.Message, file string) {
	const (
		MODULE  = "module"
		REQUIRE = "require"
		REPLACE = "replace"
		VERSION = "version"
	)

	block := ""
	require := map[string]string{}
	replace := map[string]string{}
	m.Cmd(nfs.CAT, file, func(line string) {
		ls := kit.Split(line)
		switch {
		case strings.HasPrefix(line, MODULE):
			require[ls[1]] = ""
			replace[ls[1]] = nfs.PWD
			return
		case strings.HasPrefix(line, REQUIRE+" ("):
			block = REQUIRE
			return
		case strings.HasPrefix(line, REPLACE+" ("):
			block = REPLACE
			return
		case strings.HasPrefix(line, ")"):
			block = ""
			return
		case strings.HasPrefix(line, REQUIRE):
			require[ls[1]] = ls[2]
		case strings.HasPrefix(line, REPLACE):
			replace[ls[1]] = ls[3]
		}
		if block == "" || len(ls) < 2 {
			return
		}
		switch block {
		case REQUIRE:
			require[ls[0]] = ls[1]
		case REPLACE:
			replace[ls[0]] = ls[2]
		}
	})
	for k, v := range require {
		m.Push(REQUIRE, k)
		m.Push(VERSION, v)
		m.Push(REPLACE, kit.Select("", replace[k]))
	}
	m.Sort(REPLACE)
}

const (
	TAGS = ".tags"
)
const GO = "go"
const MOD = "mod"
const SUM = "sum"
const PROTO = "proto"
const GODOC = "godoc"

func init() {
	Index.Register(&ice.Context{Name: GO, Help: "后端", Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(mdb.SEARCH, mdb.CREATE, GODOC, m.Prefix(GO))
			m.Cmd(mdb.SEARCH, mdb.CREATE, GO, m.Prefix(GO))
			m.Cmd(mdb.ENGINE, mdb.CREATE, GO, m.Prefix(GO))

			LoadPlug(m, GO, MOD, SUM, PROTO)
			for _, k := range []string{GO, MOD, SUM, PROTO, GODOC} {
				m.Cmd(mdb.RENDER, mdb.CREATE, k, m.Prefix(k))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, k, m.Prefix(k))
			}
		}},
		GODOC: {Name: "godoc", Help: "文档", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, GO, "doc", strings.TrimSuffix(arg[1], ice.PT+arg[0]), kit.Dict(cli.CMD_DIR, arg[2])).SetAppend()
			}},
		}, PlugAction())},
		PROTO: {Name: "proto", Help: "协议", Action: PlugAction()},
		SUM:   {Name: "sum", Help: "版本", Action: PlugAction()},
		MOD: {Name: "mod", Help: "模块", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _mod_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
		GO: {Name: "go", Help: "后端", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == GO {
					_go_tags(m, kit.Select(cli.MAIN, arg, 1))
					_go_help(m, kit.Select(cli.MAIN, arg, 1))
					// _go_find(m, kit.Select(cli.MAIN, arg, 1), arg[2])
					// _go_grep(m, kit.Select(cli.MAIN, arg, 1), arg[2])
				}
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _go_exec(m, arg...) }},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _go_show(m, arg...) }},
		}, PlugAction())},
	}, Configs: map[string]*ice.Config{
		PROTO: {Name: PROTO, Help: "协议", Value: kit.Data(PLUG, kit.Dict(
			PREFIX, kit.Dict("//", COMMENT), PREPARE, kit.Dict(
				KEYWORD, kit.Simple("syntax", "option", "package", "import", "service", "message"),
				DATATYPE, kit.Simple("string", "int64", "int32"),
			), KEYWORD, kit.Dict(),
		))},
		MOD: {Name: MOD, Help: "模块", Value: kit.Data(PLUG, kit.Dict(
			PREFIX, kit.Dict("//", COMMENT), PREPARE, kit.Dict(
				KEYWORD, kit.Simple("go", "module", "require", "replace", "=>"),
			), KEYWORD, kit.Dict(),
		))},
		GO: {Name: GO, Help: "后端", Value: kit.Data(PLUG, kit.Dict(
			SPLIT, kit.Dict("space", "\t ", "operator", "{[(&.,:;!|<>)]}"),
			PREFIX, kit.Dict("// ", COMMENT, "/*", COMMENT, "* ", COMMENT), PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"package", "import", "type", "struct", "interface", "const", "var", "func",
					"if", "else", "for", "range", "break", "continue",
					"switch", "case", "default", "fallthrough",
					"go", "select", "defer", "return", "panic", "recover",
				),
				DATATYPE, kit.Simple(
					"int", "int32", "int64", "float64",
					"string", "byte", "bool", "error", "chan", "map",
				),
				FUNCTION, kit.Simple(
					"init", "main", "print",
					"new", "make", "len", "cap", "copy", "append", "delete", "msg", "m",
				),
				CONSTANT, kit.Simple(
					"false", "true", "nil", "-1", "0", "1", "2",
				),
			), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
