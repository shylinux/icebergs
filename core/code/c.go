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

func _c_find(m *ice.Message, key string) {
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, FIND, ".", "-name", key), "\n") {
		if p == "" {
			continue
		}
		m.Push(kit.MDB_FILE, strings.TrimPrefix(p, "./"))
		m.Push(kit.MDB_LINE, 1)
		m.Push(kit.MDB_TEXT, "")
	}
}
func _c_grep(m *ice.Message, key string) {
	m.Split(m.Cmd(cli.SYSTEM, GREP, "--exclude-dir=.git", "--exclude-dir=pluged", "--exclude=.[a-z]*",
		"-rn", key, ".").Append(cli.CMD_OUT), "file:line:text", ":", "\n")
}
func _c_tags(m *ice.Message, key string) {
	if _, e := os.Stat(path.Join(m.Option("_path"), m.Conf(C, "meta.tags"))); e != nil {
		// 创建索引
		m.Cmd(cli.SYSTEM, CTAGS, "-R", "-f", m.Conf(C, "meta.tags"), "./")
	}

	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, GREP, "^"+key+"\\>", m.Conf(C, "meta.tags")), "\n") {
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
				m.Push(kit.MDB_FILE, strings.TrimPrefix(file, "./"))
				m.Push(kit.MDB_LINE, i)
				m.Push(kit.MDB_TEXT, bio.Text())
			}
		}
	}
	m.Sort(kit.MDB_LINE, "int")
}
func _c_help(m *ice.Message, section, key string) string {
	p := m.Cmd(cli.SYSTEM, MAN, section, key).Append(cli.CMD_OUT)
	if p == "" {
		return ""
	}

	p = strings.Replace(p, "_\x08", "", -1)
	res := make([]byte, 0, len(p))
	for i := 0; i < len(p); i++ {
		switch p[i] {
		case '\x08':
			i++
		default:
			res = append(res, p[i])
		}
	}
	return string(res)
}

const C = "c"
const H = "h"
const MAN1 = "man1"
const MAN2 = "man2"
const MAN3 = "man3"
const MAN8 = "man8"

const (
	FIND  = "find"
	GREP  = "grep"
	CTAGS = "ctags"
	MAN   = "man"
)

func init() {
	Index.Register(&ice.Context{Name: C, Help: "c",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, C, C, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, C, C, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, C, C, c.Cap(ice.CTX_FOLLOW))

				m.Cmd(mdb.SEARCH, mdb.CREATE, H, C, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, H, C, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.RENDER, mdb.CREATE, H, C, c.Cap(ice.CTX_FOLLOW))

				for _, k := range []string{MAN1, MAN2, MAN3, MAN8} {
					m.Cmd(mdb.SEARCH, mdb.CREATE, k, MAN, c.Cap(ice.CTX_FOLLOW))
					m.Cmd(mdb.PLUGIN, mdb.CREATE, k, MAN, c.Cap(ice.CTX_FOLLOW))
					m.Cmd(mdb.RENDER, mdb.CREATE, k, MAN, c.Cap(ice.CTX_FOLLOW))
				}
			}},
			C: {Name: C, Help: "c", Action: map[string]*ice.Action{
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_c_find(m, kit.Select("main", arg, 1))
					m.Cmdy(mdb.SEARCH, "man2", arg[1:])
					_c_tags(m, kit.Select("main", arg, 1))
					_c_grep(m, kit.Select("main", arg, 1))
				}},
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(C, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
			MAN: {Name: MAN, Help: "man", Action: map[string]*ice.Action{
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					for _, k := range []string{"1", "2", "3", "8"} {
						if text := _c_help(m, k, kit.Select("main", arg, 1)); text != "" {
							m.Push(kit.MDB_FILE, arg[1]+".man"+k)
							m.Push(kit.MDB_LINE, "1")
							m.Push(kit.MDB_TEXT, text)
						}
					}
				}},
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(C, "meta.man.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(_c_help(m, strings.TrimPrefix(arg[0], "man"), strings.TrimSuffix(arg[1], "."+arg[0])))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		},
		Configs: map[string]*ice.Config{
			C: {Name: C, Help: "c", Value: kit.Data(
				"tags", ".tags",
				"man.plug", kit.Dict(
					"prefix", kit.Dict(
						"NAME", "comment",
						"LIBRARY", "comment",
						"SYNOPSIS", "comment",
						"DESCRIPTION", "comment",
						"STANDARDS", "comment",
						"SEE ALSO", "comment",
						"HISTORY", "comment",
						"BUGS", "comment",
					),
				),
				"plug", kit.Dict(
					"split", kit.Dict(
						"space", " ",
						"operator", "{[(.,;!|<>)]}",
					),
					"prefix", kit.Dict(
						"//", "comment",
						"/*", "comment",
						"*", "comment",
					),
					"keyword", kit.Dict(
						"#include", "keyword",
						"#define", "keyword",
						"#ifndef", "keyword",
						"#ifdef", "keyword",
						"#else", "keyword",
						"#endif", "keyword",

						"if", "keyword",
						"else", "keyword",
						"for", "keyword",
						"while", "keyword",
						"do", "keyword",
						"break", "keyword",
						"continue", "keyword",
						"switch", "keyword",
						"case", "keyword",
						"default", "keyword",
						"return", "keyword",

						"typedef", "keyword",
						"extern", "keyword",
						"static", "keyword",
						"const", "keyword",
						"sizeof", "keyword",

						"union", "datatype",
						"struct", "datatype",
						"unsigned", "datatype",
						"double", "datatype",
						"void", "datatype",
						"long", "datatype",
						"char", "datatype",
						"int", "datatype",

						"assert", "function",
						"zmalloc", "function",

						"NULL", "string",
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
