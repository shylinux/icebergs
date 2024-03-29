package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/yac"
	kit "shylinux.com/x/toolkits"
)

func _go_trans(m *ice.Message, key string) string {
	switch key {
	case "m", "msg":
		key = "icebergs.Message"
	case "code", "wiki", "chat", "team", "mall":
		key = "shylinux.com/x/icebergs/core/" + key
	case "aaa", "cli", "ctx", "mdb", "nfs", "web":
		key = "shylinux.com/x/icebergs/base/" + key
	case "ice":
		key = "shylinux.com/x/icebergs"
	case "kit":
		key = "shylinux.com/x/toolkits"
	}
	return key
}
func _go_complete(m *ice.Message, arg ...string) {
	const (
		PACKAGE = "package"
		IMPORT  = "import"
		CONST   = "const"
		TYPE    = "type"
		FUNC    = "func"
		VAR     = "var"
	)
	if m.Option(mdb.TEXT) == "" {
		m.Push(mdb.TEXT, PACKAGE, IMPORT, CONST, TYPE, FUNC, VAR)
	} else if strings.HasSuffix(m.Option(mdb.TEXT), nfs.PT) {
		msg := m.Cmd(cli.SYSTEM, GO, "doc", _go_trans(m, kit.Slice(kit.Split(m.Option(mdb.TEXT), "\t ."), -1)[0]))
		for _, l := range strings.Split(kit.Select(msg.Result(), msg.Append(cli.CMD_OUT)), lex.NL) {
			if ls := kit.Split(l, "\t *", "()"); len(ls) > 1 {
				kit.Switch(ls[0], []string{CONST, TYPE, FUNC, VAR}, func() {
					kit.If(ls[1] == "(", func() { m.Push(mdb.NAME, ls[5]) }, func() { m.Push(mdb.NAME, ls[1]) })
					m.Push(mdb.TEXT, l)
				})
			}
		}
	} else {
		m.Push(mdb.TEXT, "m", "msg", "code", "wiki", "chat", "team", "mall", "arg", "aaa", "cli", "ctx", "mdb", "nfs", "web", "ice", "kit")
		for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, GO, "list", "std"), lex.NL) {
			m.Push(mdb.TEXT, kit.Slice(kit.Split(l, nfs.PS), -1)[0])
		}
	}
}
func _mod_show(m *ice.Message, file string) {
	const (
		MODULE  = "module"
		REQUIRE = "require"
		REPLACE = "replace"
		VERSION = "version"
	)
	require, replace, block := ice.Maps{}, ice.Maps{}, ""
	m.Cmd(nfs.CAT, file, func(ls []string, line string) {
		switch {
		case strings.HasPrefix(line, "//"):
		case strings.HasPrefix(line, MODULE):
			require[ls[1]], replace[ls[1]] = m.Cmdx(cli.SYSTEM, GIT, "describe", "--tags"), nfs.PWD
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
			kit.Switch(kit.Select("", block, len(ls) > 1), REQUIRE, func() { require[ls[0]] = ls[1] }, REPLACE, func() { replace[ls[0]] = ls[2] })
		}
	})
	kit.For(require, func(k, v string) { m.Push(REQUIRE, k).Push(VERSION, v).Push(REPLACE, kit.Select("", replace[k])) })
	m.StatusTimeCount()
}
func _sum_show(m *ice.Message, file string) {
	m.Cmdy(nfs.CAT, file, func(ls []string, line string) {
		m.Push(nfs.REPOS, ls[0]).Push(nfs.VERSION, ls[1]).Push(mdb.HASH, ls[2])
	}).StatusTimeCount()
}

const SUM = "sum"
const MOD = "mod"
const GO = "go"

func init() {
	Index.MergeCommands(ice.Commands{
		"godoc": {Name: "godoc key auto", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(cli.SYSTEM, "go", "list", "std")
			} else {
				m.Cmdy(cli.SYSTEM, "go", "doc", arg)
			}
		}},
		GO: {Name: "go path auto", Help: "后端编程", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				if arg[1] == "main.go" {
					ProcessXterm(m, "ish", "", arg[1])
				} else if arg[1] == "misc/xterm/iterm.go" {
					ProcessXterm(m, "ish", "", arg[1])
				} else if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
					ctx.ProcessCommand(m, cmd, kit.Simple())
				} else if msg := m.Cmd(yac.STACK, path.Join(arg[2], arg[1])); msg.Option("__index") != "" {
					ctx.ProcessCommand(m, msg.Option("__index"), kit.Simple())
				} else {
					ctx.ProcessCommand(m, yac.STACK, kit.Simple(path.Join(arg[2], arg[1])))
				}
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				if arg[1] == "main.go" {
					ProcessXterm(m, "ish", "", arg[1])
				} else if arg[1] == "misc/xterm/iterm.go" {
					ProcessXterm(m, "ish", "", arg[1])
				} else if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
					ctx.ProcessCommand(m, cmd, kit.Simple())
				} else if msg := m.Cmd(yac.STACK, path.Join(arg[2], arg[1])); msg.Option("__index") != "" {
					ctx.ProcessCommand(m, msg.Option("__index"), kit.Simple())
				} else {
					ctx.ProcessCommand(m, yac.STACK, kit.Simple(path.Join(arg[2], arg[1])))
				}
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Option("name", kit.TrimExt(path.Base(arg[1]), "go"))
				m.Option("zone", path.Base(path.Dir(path.Join(arg[2], arg[1]))))
				m.Option("key", kit.Keys("web.code", m.Option("zone"), m.Option("name")))
				m.Echo(nfs.Template(m, "demo.go"))
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) { _go_complete(m, arg...) }},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{"guru", "gopls"} {
					if ls := kit.Split(m.Cmdx(cli.SYSTEM, cmd, "definition", m.Option(nfs.PATH)+m.Option(nfs.FILE)+nfs.DF+"#"+m.Option("offset")), nfs.DF); len(ls) > 0 {
						if strings.HasPrefix(ls[0], kit.Path("")) {
							_ls := nfs.SplitPath(m, strings.TrimPrefix(ls[0], kit.Path("")+nfs.PS))
							m.Push(nfs.PATH, _ls[0]).Push(nfs.FILE, _ls[1]).Push(nfs.LINE, ls[1])
							return
						}
					}
				}
				_c_tags(m, "gotags", "-f", nfs.TAGS, "-R", nfs.PWD)
			}},
		}, PlugAction())},
		MOD: {Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _mod_show(m, path.Join(arg[2], arg[1])) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _mod_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
		SUM: {Actions: ice.MergeActions(ice.Actions{
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _sum_show(m, path.Join(arg[2], arg[1])) }},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _sum_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
	})
}
