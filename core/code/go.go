package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
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
	} else if strings.HasSuffix(m.Option(mdb.TEXT), ice.PT) {
		msg := m.Cmd(cli.SYSTEM, GO, "doc", _go_trans(m, kit.Slice(kit.Split(m.Option(mdb.TEXT), "\t ."), -1)[0]))
		for _, l := range strings.Split(kit.Select(msg.Result(), msg.Append(cli.CMD_OUT)), ice.NL) {
			if ls := kit.Split(l, "\t *", "()"); len(ls) > 1 {
				kit.Switch(ls[0], []string{CONST, TYPE, FUNC, VAR}, func() {
					kit.If(ls[1] == "(", func() { m.Push(mdb.NAME, ls[5]) }, func() { m.Push(mdb.NAME, ls[1]) })
					m.Push(mdb.TEXT, l)
				})
			}
		}
	} else {
		m.Push(mdb.TEXT, "m", "msg", "code", "wiki", "chat", "team", "mall", "arg", "aaa", "cli", "ctx", "mdb", "nfs", "web", "ice", "kit")
		for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, GO, "list", "std"), ice.NL) {
			m.Push(mdb.TEXT, kit.Slice(kit.Split(l, ice.PS), -1)[0])
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
	kit.Fetch(require, func(k, v string) { m.Push(REQUIRE, k).Push(VERSION, v).Push(REPLACE, kit.Select("", replace[k])) })
	m.StatusTimeCount()
}
func _sum_show(m *ice.Message, file string) {
	m.Cmdy(nfs.CAT, file, func(ls []string, line string) {
		m.Push(nfs.REPOS, ls[0]).Push(nfs.VERSION, ls[1]).Push(mdb.HASH, ls[2])
	}).StatusTimeCount()
}

const GODOC = "godoc"
const SUM = "sum"
const MOD = "mod"
const GO = "go"

func init() {
	Index.MergeCommands(ice.Commands{
		GO: {Name: "go path auto", Help: "后端编程", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(NAVIGATE, mdb.CREATE, GODOC, m.PrefixKey()) }},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				cmds, text := "ice.bin source stdio", ctx.GetFileCmd(path.Join(arg[2], arg[1]))
				if text != "" {
					ls := strings.Split(text, ice.PT)
					text = "~" + kit.Join(kit.Slice(ls, 0, -1), ice.PT) + ice.NL + kit.Slice(ls, -1)[0]
				} else {
					text = "cli.system go run " + path.Join(arg[2], arg[1])
				}
				_xterm_show(m, cmds, text)
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
					ctx.ProcessCommand(m, cmd, kit.Simple())
				} else {
					cmds := []string{GO, ice.RUN, path.Join(arg[2], arg[1])}
					m.Cmdy(cli.SYSTEM, cmds).StatusTime(ssh.SHELL, strings.Join(cmds, ice.SP))
				}
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_go_template, path.Base(path.Dir(path.Join(arg[2], arg[1]))))
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) { _go_complete(m, arg...) }},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, GODOC, "gotags", "-f", nfs.TAGS, "-R", nfs.PWD) }},
		}, PlugAction())},
		MOD: {Help: "模块", Actions: ice.MergeActions(ice.Actions{
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _mod_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
		SUM: {Help: "版本", Actions: ice.MergeActions(ice.Actions{
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _sum_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
		GODOC: {Help: "文档", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				arg[1] = strings.Replace(arg[1], "kit.", "shylinux.com/x/toolkits.", 1)
				arg[1] = strings.Replace(arg[1], "m.", "shylinux.com/x/ice.Message.", 1)
				m.Cmdy(cli.SYSTEM, GO, "doc", kit.TrimExt(arg[1], GODOC), kit.Dict(cli.CMD_DIR, arg[2]))
			}},
		}, PlugAction())},
	})
}

var _go_template = `package %s

func init() {
	
}
`
