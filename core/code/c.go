package code

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _c_tags(m *ice.Message, key string) {
	if _, e := os.Stat(path.Join(m.Option(cli.CMD_DIR), _TAGS)); e != nil {
		m.Cmd(cli.SYSTEM, "ctags", "-R", "-f", _TAGS, "./")
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

const (
	H    = "h"
	CC   = "cc"
	MAN  = "man"
	MAN1 = "man1"
	MAN2 = "man2"
	MAN3 = "man3"
	MAN8 = "man8"
)
const C = "c"

func init() {
	Index.Register(&ice.Context{Name: C, Help: "系统", Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for _, cmd := range []string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE, mdb.SEARCH} {
				for _, k := range []string{H, C, CC} {
					m.Cmd(cmd, mdb.CREATE, k, m.Prefix(C))
				}
			}
			for _, cmd := range []string{mdb.PLUGIN, mdb.RENDER, mdb.SEARCH} {
				for _, k := range []string{MAN1, MAN2, MAN3, MAN8} {
					m.Cmd(cmd, mdb.CREATE, k, m.Prefix(MAN))
				}
			}
			LoadPlug(m, C)
			LoadPlug(m, MAN)
		}},
		C: {Name: C, Help: "系统", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, arg[2])
				name := strings.TrimSuffix(arg[1], path.Ext(arg[1])) + ".bin"
				if msg := m.Cmd(cli.SYSTEM, "gcc", arg[1], "-o", name); !cli.IsSuccess(msg) {
					m.Copy(msg)
					return
				}
				m.Cmdy(cli.SYSTEM, "./"+name)
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == kit.MDB_FOREACH {
					return
				}
				m.Option(cli.CMD_DIR, kit.Select(ice.SRC, arg, 2))
				m.Cmdy(mdb.SEARCH, MAN2, arg[1:])
				m.Cmdy(mdb.SEARCH, MAN3, arg[1:])
				_c_tags(m, kit.Select(kit.MDB_MAIN, arg, 1))
				_go_find(m, kit.Select(kit.MDB_MAIN, arg, 1))
				_go_grep(m, kit.Select(kit.MDB_MAIN, arg, 1))
			}},
		}, PlugAction())},
		MAN: {Name: MAN, Help: "手册", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_c_help(m, strings.TrimPrefix(arg[0], MAN), strings.TrimSuffix(arg[1], ice.PT+arg[0])))
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == kit.MDB_FOREACH {
					return
				}
				for _, i := range []string{"1", "2", "3", "8"} {
					if text := _c_help(m, i, kit.Select(kit.MDB_MAIN, arg, 1)); text != "" {
						m.PushSearch(ice.CMD, MAN, kit.MDB_FILE, kit.Keys(arg[1], MAN+i), kit.MDB_LINE, 1, kit.MDB_TEXT, text)
					}
				}
			}},
		}, PlugAction())},
	}, Configs: map[string]*ice.Config{
		C: {Name: C, Help: "系统", Value: kit.Data(PLUG, kit.Dict(
			SPLIT, kit.Dict("space", " ", "operator", "{[(.,:;!|<>)]}"),
			PREFIX, kit.Dict("//", COMMENT, "/*", COMMENT, "*", COMMENT),
			PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"#include",
					"#define",
					"#ifndef",
					"#ifdef",
					"#else",
					"#endif",

					"if",
					"else",
					"for",
					"while",
					"do",
					"break",
					"continue",
					"switch",
					"case",
					"default",
					"return",

					"typedef",
					"sizeof",
					"extern",
					"static",
					"const",
				),
				DATATYPE, kit.Simple(
					"union",
					"struct",
					"unsigned",
					"double",
					"void",
					"long",
					"char",
					"int",
				),
				FUNCTION, kit.Simple(
					"assert",
					"zmalloc",
				),
				CONSTANT, kit.Simple(
					"NULL", "-1", "0", "1", "2",
				),
			), KEYWORD, kit.Dict(),
		))},
		MAN: {Name: MAN, Help: "手册", Value: kit.Data(PLUG, kit.Dict(
			PREPARE, kit.Dict(
				COMMENT, kit.Simple(
					"NAME",
					"LIBRARY",
					"SYNOPSIS",
					"DESCRIPTION",
					"STANDARDS",
					"SEE ALSO",
					"HISTORY",
					"BUGS",
				),
			), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
