package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

func _c_tags(m *ice.Message, key string) {
	if _, e := os.Stat(path.Join(m.Option(cli.CMD_DIR), ".tags")); e != nil {
		m.Cmd(cli.SYSTEM, "ctags", "-R", "-f", ".tags", "./")
	}
	_go_tags(m, key)
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

const H = "h"
const C = "c"
const CC = "cc"
const MAN1 = "man1"
const MAN2 = "man2"
const MAN3 = "man3"
const MAN8 = "man8"

const (
	FIND = "find"
	GREP = "grep"
	MAN  = "man"
)

func init() {
	Index.Register(&ice.Context{Name: C, Help: "系统",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.PLUGIN, mdb.CREATE, CC, m.Prefix(C))
				m.Cmd(mdb.RENDER, mdb.CREATE, CC, m.Prefix(C))
				m.Cmd(mdb.SEARCH, mdb.CREATE, CC, m.Prefix(C))

				m.Cmd(mdb.PLUGIN, mdb.CREATE, C, m.Prefix(C))
				m.Cmd(mdb.RENDER, mdb.CREATE, C, m.Prefix(C))
				m.Cmd(mdb.SEARCH, mdb.CREATE, C, m.Prefix(C))

				m.Cmd(mdb.PLUGIN, mdb.CREATE, H, m.Prefix(C))
				m.Cmd(mdb.RENDER, mdb.CREATE, H, m.Prefix(C))
				m.Cmd(mdb.SEARCH, mdb.CREATE, H, m.Prefix(C))

				for _, k := range []string{MAN1, MAN2, MAN3, MAN8} {
					m.Cmd(mdb.PLUGIN, mdb.CREATE, k, m.Prefix(MAN))
					m.Cmd(mdb.RENDER, mdb.CREATE, k, m.Prefix(MAN))
					m.Cmd(mdb.SEARCH, mdb.CREATE, k, m.Prefix(MAN))
				}
			}},
			C: {Name: C, Help: "系统", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(C, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				}},
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					m.Option(cli.CMD_DIR, kit.Select("src", arg, 2))
					_go_find(m, kit.Select("main", arg, 1))
					m.Cmdy(mdb.SEARCH, MAN2, arg[1:])
					m.Cmdy(mdb.SEARCH, MAN3, arg[1:])
					_c_tags(m, kit.Select("main", arg, 1))
					_go_grep(m, kit.Select("main", arg, 1))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
			MAN: {Name: MAN, Help: "手册", Action: map[string]*ice.Action{
				mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Conf(MAN, "meta.plug"))
				}},
				mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
					m.Echo(_c_help(m, strings.TrimPrefix(arg[0], MAN), strings.TrimSuffix(arg[1], "."+arg[0])))
				}},
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_FOREACH {
						return
					}
					for _, i := range []string{"1", "2", "3", "8"} {
						if text := _c_help(m, i, kit.Select("main", arg, 1)); text != "" {
							for _, k := range kit.Split(m.Option(mdb.FIELDS)) {
								switch k {
								case kit.MDB_FILE:
									m.Push(k, arg[1]+".man"+i)
								case kit.MDB_LINE:
									m.Push(k, 1)
								case kit.MDB_TEXT:
									m.Push(k, text)
								default:
									m.Push(k, "")
								}
							}
						}
					}
				}},
			}},
		},
		Configs: map[string]*ice.Config{
			MAN: {Name: MAN, Help: "手册", Value: kit.Data(
				"plug", kit.Dict(
					PREFIX, kit.Dict(
						"NAME", COMMENT,
						"LIBRARY", COMMENT,
						"SYNOPSIS", COMMENT,
						"DESCRIPTION", COMMENT,
						"STANDARDS", COMMENT,
						"SEE ALSO", COMMENT,
						"HISTORY", COMMENT,
						"BUGS", COMMENT,
					),
				),
			)},
			C: {Name: C, Help: "系统", Value: kit.Data(
				"plug", kit.Dict(
					SPLIT, kit.Dict(
						"space", " ",
						"operator", "{[(.,:;!|<>)]}",
					),
					PREFIX, kit.Dict(
						"//", COMMENT,
						"/*", COMMENT,
						"*", COMMENT,
					),
					KEYWORD, kit.Dict(
						"#include", KEYWORD,
						"#define", KEYWORD,
						"#ifndef", KEYWORD,
						"#ifdef", KEYWORD,
						"#else", KEYWORD,
						"#endif", KEYWORD,

						"if", KEYWORD,
						"else", KEYWORD,
						"for", KEYWORD,
						"while", KEYWORD,
						"do", KEYWORD,
						"break", KEYWORD,
						"continue", KEYWORD,
						"switch", KEYWORD,
						"case", KEYWORD,
						"default", KEYWORD,
						"return", KEYWORD,

						"typedef", KEYWORD,
						"extern", KEYWORD,
						"static", KEYWORD,
						"const", KEYWORD,
						"sizeof", KEYWORD,

						"union", DATATYPE,
						"struct", DATATYPE,
						"unsigned", DATATYPE,
						"double", DATATYPE,
						"void", DATATYPE,
						"long", DATATYPE,
						"char", DATATYPE,
						"int", DATATYPE,

						"assert", FUNCTION,
						"zmalloc", FUNCTION,

						"NULL", STRING,
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
