package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
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
	if m.Option(mdb.TEXT) == "" {
		m.Push(mdb.TEXT, PACKAGE, IMPORT, CONST, TYPE, FUNC, VAR)
	} else if strings.HasSuffix(m.Option(mdb.TEXT), nfs.PT) {
		for _, l := range strings.Split(GoDoc(m, _go_trans(m, kit.Slice(kit.Split(m.Option(mdb.TEXT), "\t ."), -1)[0])), lex.NL) {
			if ls := kit.Split(l, "\t *", "()"); len(ls) > 1 {
				kit.Switch(ls[0], []string{CONST, TYPE, FUNC, VAR}, func() {
					kit.If(ls[1] == "(", func() { m.Push(mdb.NAME, ls[5]) }, func() { m.Push(mdb.NAME, ls[1]) })
					m.Push(mdb.TEXT, l)
				})
			}
		}
	} else {
		m.Push(mdb.TEXT, "m", "msg", "code", "wiki", "chat", "team", "mall", "arg", "aaa", "cli", "ctx", "mdb", "nfs", "web", "ice", "kit")
		for _, l := range strings.Split(GoStd(m), lex.NL) {
			m.Push(mdb.TEXT, kit.Slice(kit.Split(l, nfs.PS), -1)[0])
		}
	}
}
func _go_navigate(m *ice.Message, arg ...string) {
	for _, cmd := range []string{"guru", "gopls"} {
		if ls := kit.Split(m.Cmdx(cli.SYSTEM, cmd, "definition", m.Option(nfs.PATH)+m.Option(nfs.FILE)+nfs.DF+"#"+m.Option("offset")), nfs.DF); len(ls) > 0 {
			if strings.HasPrefix(ls[0], kit.Path("")) {
				_ls := nfs.SplitPath(m, strings.TrimPrefix(ls[0], kit.Path("")+nfs.PS))
				m.Push(nfs.PATH, _ls[0]).Push(nfs.FILE, _ls[1]).Push(nfs.LINE, ls[1])
				return
			}
		}
	}
	_c_tags(m, GOTAGS, "-f", nfs.TAGS, "-R", nfs.PWD)
}
func _go_show(m *ice.Message, arg ...string) {
	if arg[1] == MAIN_GO {
		ProcessXterm(m, kit.JoinWord("ish", arg[1]), "", "")
	} else if arg[1] == "version.go" {
		ctx.ProcessField(m, cli.RUNTIME, kit.Simple())
	} else if arg[1] == "binpack.go" {
		ctx.ProcessField(m, nfs.PACK, kit.Simple())
	} else if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
		ctx.ProcessField(m, cmd, kit.Simple())
	} else if msg := m.Cmd(yac.STACK, path.Join(arg[2], arg[1])); msg.Option("__index") != "" {
		ctx.ProcessField(m, msg.Option("__index"), kit.Simple())
	} else {
		ctx.ProcessField(m, yac.STACK, kit.Simple(path.Join(arg[2], arg[1])))
	}
}
func _mod_show(m *ice.Message, file string) {
	require, replace, block := ice.Maps{}, ice.Maps{}, ""
	m.Cmd(nfs.CAT, file, func(ls []string, line string) {
		switch {
		case strings.HasPrefix(line, "//"):
		case strings.HasPrefix(line, MODULE):
			require[ls[1]], replace[ls[1]] = ice.Info.Make.Versions(), nfs.PWD
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
}
func _sum_show(m *ice.Message, file string) {
	m.Cmdy(nfs.CAT, file, func(ls []string, line string) {
		m.Push(nfs.REPOS, ls[0]).Push(nfs.VERSION, ls[1]).Push(mdb.HASH, ls[2])
	})
}

const (
	// PACKAGE = "package"
	IMPORT = "import"
	CONST  = "const"
	TYPE   = "type"
	FUNC   = "func"
	VAR    = "var"
)
const (
	MODULE  = "module"
	REQUIRE = "require"
	REPLACE = "replace"
	VERSION = "version"
	SERVICE = "service"
)
const (
	GOFMT     = "gofmt"
	GOIMPORTS = "goimports"
	GOTAGS    = "gotags"
)
const (
	GOPRIVATE   = "GOPRIVATE"
	GOPROXY     = "GOPROXY"
	GOPATH      = "GOPATH"
	GOCACHE     = "GOCACHE"
	GOMODCACHE  = "GOMODCACHE"
	CGO_ENABLED = "CGO_ENABLED"
)
const (
	ENV = "env"
	GET = "get"
	TAG = "tag"
)
const SUM = "sum"
const MOD = "mod"
const DOC = "doc"
const GO = "go"

func init() {
	Index.MergeCommands(ice.Commands{
		GO: {Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { cli.IsRedhat(m, GO, GO); cli.IsAlpine(m, GO, GO) }},
			mdb.RENDER:   {Hand: func(m *ice.Message, arg ...string) { _go_show(m, arg...) }},
			mdb.ENGINE:   {Hand: func(m *ice.Message, arg ...string) { _go_show(m, arg...) }},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.NAME, kit.TrimExt(path.Base(arg[1]), GO))
				m.Option(mdb.ZONE, path.Base(path.Dir(path.Join(arg[2], arg[1]))))
				m.Option(mdb.KEY, Prefix(m.Option(mdb.ZONE), m.Option(mdb.NAME)))
				m.Echo(nfs.Template(m, DEMO_GO))
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) { _go_complete(m, arg...) }},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) { _go_navigate(m, arg...) }},
			VIMER_SAVE: {Hand: func(m *ice.Message, arg ...string) {
				switch p := path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)); kit.Ext(m.Option(nfs.FILE)) {
				case nfs.GO:
					defer m.Cmdy(nfs.CAT, p)
					defer GoImports(m, p)
					defer GoFmt(m, p)
				}
			}},
		}, gdb.EventsAction(VIMER_SAVE), PlugAction())},
		DOC: {Name: "doc path name auto", Hand: func(m *ice.Message, arg ...string) {
			if strings.Contains(arg[0], nfs.PS) {
				arg[0] = kit.Path(arg[0])
			}
			m.Cmdy(cli.SYSTEM, cli.GO, DOC, arg)
		}},
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
func GoEnv(m *ice.Message, k string) string { return m.Cmdx(cli.SYSTEM, GO, ENV, k) }
func GoDoc(m *ice.Message, k string) string {
	msg := m.Cmd(cli.SYSTEM, GO, DOC, k)
	return kit.Select(msg.Result(), msg.Append(cli.CMD_OUT))
}
func GoStd(m *ice.Message) string        { return m.Cmdx(cli.SYSTEM, GO, "list", "std") }
func GoGet(m *ice.Message, p string)     { m.Cmd(cli.SYSTEM, GO, GET, p) }
func GoFmt(m *ice.Message, p string)     { m.Cmd(cli.SYSTEM, GOFMT, "-w", p) }
func GoImports(m *ice.Message, p string) { m.Cmd(cli.SYSTEM, GOIMPORTS, "-w", p) }
func GoVersion(m *ice.Message) string    { return m.Cmdx(cli.SYSTEM, GO, VERSION) }
func GoBuild(m *ice.Message, target string, arg ...string) *ice.Message {
	return m.Cmdy(cli.SYSTEM, GO, cli.BUILD, "-ldflags", "-w -s", "-o", target, arg)
}
