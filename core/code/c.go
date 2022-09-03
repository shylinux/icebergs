package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _c_tags(m *ice.Message, key string) {
	if !nfs.ExistsFile(m, path.Join(m.Option(cli.CMD_DIR), TAGS)) {
		m.Cmd(cli.SYSTEM, "ctags", "-R", "-f", TAGS, nfs.PWD)
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
	Index.Register(&ice.Context{Name: C, Help: "系统", Commands: ice.Commands{
		C: {Name: C, Help: "系统", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{mdb.SEARCH, mdb.ENGINE, mdb.RENDER, mdb.PLUGIN} {
					for _, k := range []string{H, C, CC} {
						m.Cmd(cmd, mdb.CREATE, k, m.PrefixKey())
					}
				}
				LoadPlug(m, H, C)
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				name := strings.TrimSuffix(arg[1], path.Ext(arg[1])) + ".bin"
				if msg := m.Cmd(cli.SYSTEM, "gcc", arg[1], "-o", name, kit.Dict(cli.CMD_DIR, arg[2])); !cli.IsSuccess(msg) {
					m.Copy(msg)
					return
				}
				m.Echo(m.Cmd(cli.SYSTEM, path.Join(arg[2], name)).Append(cli.CMD_OUT))
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				TagsList(m, "ctags", "--excmd=number", "--sort=no", "-f", "-", path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH {
					return
				}
				m.Option(cli.CMD_DIR, kit.Select(ice.SRC, arg, 2))
				m.Cmdy(mdb.SEARCH, MAN2, arg[1:])
				m.Cmdy(mdb.SEARCH, MAN3, arg[1:])
				_c_tags(m, kit.Select(cli.MAIN, arg, 1))
				// _go_find(m, kit.Select(cli.MAIN, arg, 1), arg[2])
				// _go_grep(m, kit.Select(cli.MAIN, arg, 1), arg[2])
			}},
		}, PlugAction())},
		MAN: {Name: MAN, Help: "手册", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{mdb.SEARCH, mdb.RENDER, mdb.PLUGIN} {
					for _, k := range []string{MAN1, MAN2, MAN3, MAN8} {
						m.Cmd(cmd, mdb.CREATE, k, m.PrefixKey())
					}
				}
				LoadPlug(m, MAN)
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_c_help(m, strings.TrimPrefix(arg[0], MAN), strings.TrimSuffix(arg[1], ice.PT+arg[0])))
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH {
					return
				}
				for _, i := range []string{"1", "2", "3", "8"} {
					if text := _c_help(m, i, kit.Select(cli.MAIN, arg, 1)); text != "" {
						m.PushSearch(ice.CMD, MAN, nfs.FILE, kit.Keys(arg[1], MAN+i), nfs.LINE, 1, mdb.TEXT, text)
					}
				}
			}},
		}, PlugAction())},
	}, Configs: ice.Configs{
		C: {Name: C, Help: "系统", Value: kit.Data(PLUG, kit.Dict(
			mdb.RENDER, kit.Dict(),
			SPLIT, kit.Dict("space", " ", "operator", "{[(.,:;!|<>)]}"),
			PREFIX, kit.Dict("//", COMMENT, "/* ", COMMENT, "* ", COMMENT), PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"#include",
					"#define",
					"#ifndef",
					"#ifdef",
					"#if",
					"#elif",
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

func TagsList(m *ice.Message, cmds ...string) {
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, cmds), ice.NL) {
		if strings.HasPrefix(l, "!_") {
			continue
		}
		ls := strings.Split(l, ice.TB)
		if len(ls) < 2 {
			continue
		}
		switch ls[3] {
		case "w":
			continue
		}
		m.PushRecord(kit.Dict(mdb.TYPE, ls[3], mdb.NAME, ls[0], nfs.LINE, strings.TrimSuffix(ls[2], ";\"")))
	}
	m.Sort(nfs.LINE).StatusTimeCount()
}
