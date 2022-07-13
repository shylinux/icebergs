package code

import (
	"bufio"
	"os"
	"path"
	"strings"
	"sync"
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
	m.Cmd(nfs.FIND, dir, key).Tables(func(value ice.Maps) { m.PushSearch(nfs.LINE, 1, value) })
}
func _go_grep(m *ice.Message, key string, dir string) {
	m.Cmd(nfs.GREP, dir, key).Tables(func(value ice.Maps) { m.PushSearch(value) })
}

var _cache_mods = map[string]*ice.Message{}
var _cache_lock = sync.Mutex{}

func _go_doc(m *ice.Message, mod string, pkg string) *ice.Message {
	_cache_lock.Lock()
	defer _cache_lock.Unlock()

	key := kit.Keys(mod, pkg)
	if msg, ok := _cache_mods[key]; ok && kit.Time(msg.Time("24h")) > kit.Time(m.Time()) {
		return msg
	}

	if mod != "" {
		m.Cmd(cli.SYSTEM, "go", "get", mod)
	}
	if msg := _vimer_go_complete(m.Spawn(), key); msg.Length() > 0 {
		_cache_mods[key] = msg
		return msg
	}
	return nil
}

func _go_exec(m *ice.Message, arg ...string) {
	if m.Option("some") == "run" {
		m.Cmdy(cli.SYSTEM, "./bin/ice.bin", ice.GetFileCmd(path.Join(arg[2], arg[1])))
		return
	}
	if m.Option(mdb.TEXT) == "" {
		if m.Option(nfs.LINE) == "1" {
			m.Push(mdb.NAME, "package")
		} else {
			m.Push(mdb.NAME, "import")
			m.Push(mdb.NAME, "const")
			m.Push(mdb.NAME, "type")
			m.Push(mdb.NAME, "func")
		}
		return
	}

	if m.Option(mdb.NAME) == ice.PT {
		switch m.Option(mdb.TYPE) {
		case "msg", "m":
			m.Copy(_go_doc(m, "shylinux.com/x/ice", "Message"))
			m.Copy(_go_doc(m, "shylinux.com/x/icebergs", "Message"))

		case "ice", "*ice":
			m.Copy(_go_doc(m, "shylinux.com/x/ice", ""))

		case "kit":
			m.Copy(_go_doc(m, "shylinux.com/x/toolkits", ""))

		default:
			m.Copy(_go_doc(m, "", m.Option(mdb.TYPE)))
		}

	} else {
		m.Push(mdb.NAME, "msg")
		m.Push(mdb.NAME, "ice")
	}
}
func _go_show(m *ice.Message, arg ...string) {
	if arg[1] == "main.go" {
		const (
			PACKAGE = "package"
			IMPORT  = "import"
		)
		index := 0
		push := func(repos string) {
			index++
			m.Push("index", index)
			m.Push("repos", repos)
		}
		block := ""
		m.Cmd(nfs.CAT, path.Join(arg[2], arg[1]), func(ls []string, line string) {
			switch {
			case strings.HasPrefix(line, IMPORT+" ("):
				block = IMPORT
			case strings.HasPrefix(line, ")"):
				block = ""
			case strings.HasPrefix(line, IMPORT):
				if len(ls) == 2 {
					push(ls[1])
				} else if len(ls) == 3 {
					push(ls[2])
				}
			default:
				if block == IMPORT {
					if len(ls) == 0 {
						push("")
					} else if len(ls) == 1 {
						push(ls[0])
					} else if len(ls) == 2 {
						push(ls[1])
					}
				}
			}
		})
	} else {
		if key := ice.GetFileCmd(path.Join(arg[2], arg[1])); key != "" {
			m.ProcessCommand(key, kit.Simple())
		} else {
			m.ProcessCommand("web.wiki.word", kit.Simple(strings.ReplaceAll(path.Join(arg[2], arg[1]), ".go", ".shy")))
		}
	}
}
func _sum_show(m *ice.Message, file string) {
	m.Cmd(nfs.CAT, file, func(ls []string, line string) {
		m.Push("repos", ls[0])
		m.Push("version", ls[1])
		m.Push("hash", ls[2])
	})
	m.StatusTimeCount()
}
func _mod_show(m *ice.Message, file string) {
	const (
		MODULE  = "module"
		REQUIRE = "require"
		REPLACE = "replace"
		VERSION = "version"
	)

	block := ""
	require := ice.Maps{}
	replace := ice.Maps{}
	m.Cmd(nfs.CAT, file, func(ls []string, line string) {
		switch {
		case strings.HasPrefix(line, MODULE):
			require[ls[1]] = m.Cmdx(cli.SYSTEM, GIT, "describe", "--tags")
			replace[ls[1]] = nfs.PWD
		case strings.HasPrefix(line, REQUIRE+" ("):
			block = REQUIRE
		case strings.HasPrefix(line, REPLACE+" ("):
			block = REPLACE
		case strings.HasPrefix(line, ")"):
			block = ""
		case strings.HasPrefix(line, REQUIRE):
			require[ls[1]] = ls[2]
		case strings.HasPrefix(line, REPLACE):
			replace[ls[1]] = ls[3]
		default:
			if len(ls) > 1 {
				switch block {
				case REQUIRE:
					require[ls[0]] = ls[1]
				case REPLACE:
					replace[ls[0]] = ls[2]
				}
			}
		}
	})
	for k, v := range require {
		m.Push(REQUIRE, k)
		m.Push(VERSION, v)
		m.Push(REPLACE, kit.Select("", replace[k]))
	}
	m.Sort(REPLACE)
	m.StatusTimeCount()
}

const (
	TAGS = ".tags"
)
const GO = "go"
const MOD = "mod"
const SUM = "sum"
const GODOC = "godoc"

func init() {
	Index.Register(&ice.Context{Name: GO, Help: "后端", Commands: ice.Commands{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(mdb.SEARCH, mdb.CREATE, GODOC, m.Prefix(GO))
			m.Cmd(mdb.ENGINE, mdb.CREATE, GO, m.Prefix(GO))

			LoadPlug(m, GO, MOD, SUM)
			for _, k := range []string{GO, MOD, SUM, GODOC} {
				m.Cmd(mdb.RENDER, mdb.CREATE, k, m.Prefix(k))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, k, m.Prefix(k))
			}
		}},
		GODOC: {Name: "godoc", Help: "文档", Actions: ice.MergeAction(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, GO, "doc", strings.TrimSuffix(arg[1], ice.PT+arg[0]), kit.Dict(cli.CMD_DIR, arg[2])).SetAppend()
			}},
		}, PlugAction())},
		SUM: {Name: "sum", Help: "版本", Actions: ice.MergeAction(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _sum_show(m, path.Join(arg[2], arg[1])) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _sum_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
		MOD: {Name: "mod", Help: "模块", Actions: ice.MergeAction(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _mod_show(m, path.Join(arg[2], arg[1])) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _mod_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
		GO: {Name: "go", Help: "后端", Actions: ice.MergeAction(ice.Actions{
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
	}, Configs: ice.Configs{
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
					"go", "select", "defer", "return",
				),
				CONSTANT, kit.Simple(
					"false", "true", "nil", "iota", "-1", "0", "1", "2", "3",
				),
				DATATYPE, kit.Simple(
					"int", "int8", "int16", "int32", "int64",
					"uint", "uint8", "uint16", "uint32", "uint64",
					"float32", "float64", "complex64", "complex128",
					"rune", "string", "byte", "uintptr",
					"bool", "error", "chan", "map",
				),
				FUNCTION, kit.Simple("msg", "m",
					"init", "main", "print", "println", "panic", "recover",
					"new", "make", "len", "cap", "copy", "append", "delete", "close",
					"complex", "real", "imag",
				),
			), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
