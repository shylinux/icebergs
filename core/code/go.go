package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _go_trans(m *ice.Message, key string) string {
	switch key {
	case "m", "msg":
		key = "icebergs.Message"
	case "kit":
		key = "shylinux.com/x/toolkits"
	case "ice":
		key = "shylinux.com/x/ice"
	case "mdb", "cli", "nfs":
		key = "shylinux.com/x/icebergs/base/" + key
	}
	return key
}
func _go_complete(m *ice.Message, arg ...string) {
	if m.Option(mdb.TEXT) == "" {
		m.Push(mdb.TEXT, "package", "import", "const", "type", "func", "var")
		return
	}

	if strings.HasSuffix(m.Option(mdb.TEXT), ice.PT) {
		key := kit.Slice(kit.Split(m.Option(mdb.TEXT), "\t ."), -1)[0]
		key = _go_trans(m, key)

		msg := m.Cmd(cli.SYSTEM, GO, "doc", key)
		for _, l := range strings.Split(kit.Select(msg.Result(), msg.Append(cli.CMD_OUT)), ice.NL) {
			ls := kit.Split(l, "\t *", "()")
			if len(ls) < 2 {
				continue
			}
			switch ls[0] {
			case "const", "type", "func", "var":
				if ls[1] == "(" {
					m.Push(mdb.NAME, ls[5])
				} else {
					m.Push(mdb.NAME, ls[1])
				}
				m.Push(mdb.TEXT, l)
			}
		}
		return
	}

	m.Push(mdb.TEXT, "m", "msg", "arg", "mdb", "cli", "nfs", "ice", "kit")
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, GO, "list", "std"), ice.NL) {
		m.Push(mdb.TEXT, kit.Slice(kit.Split(l, ice.PS), -1)[0])
	}
}
func _go_exec(m *ice.Message, arg ...string) {
	if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
		ctx.ProcessCommand(m, cmd, kit.Simple())
		return
	}
}
func _go_show(m *ice.Message, arg ...string) {
	TagsList(m, "gotags", path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
	return
	if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
		ctx.ProcessCommand(m, cmd, kit.Simple())
	} else if p := strings.ReplaceAll(path.Join(arg[2], arg[1]), ".go", ".shy"); arg[1] != "main.go" && nfs.ExistsFile(m, p) {
		ctx.ProcessCommand(m, "web.wiki.word", kit.Simple(p))
	} else {
		TagsList(m, "gotags", path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
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
	require := ice.Maps{}
	replace := ice.Maps{}
	m.Cmd(nfs.CAT, file, func(ls []string, line string) {
		switch {
		case strings.HasPrefix(line, "//"):
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
	m.Sort(REPLACE).StatusTimeCount()
}
func _sum_show(m *ice.Message, file string) {
	m.Cmd(nfs.CAT, file, func(ls []string, line string) {
		m.Push("repos", ls[0])
		m.Push("version", ls[1])
		m.Push("hash", ls[2])
	})
	m.StatusTimeCount()
}

const GO = "go"
const GODOC = "godoc"
const MOD = "mod"
const SUM = "sum"

func init() {
	Index.MergeCommands(ice.Commands{
		GO: {Name: "go path auto", Help: "后端", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(NAVIGATE, mdb.CREATE, GODOC, m.PrefixKey())
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _go_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _go_exec(m, arg...) }},

			COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == mdb.FOREACH {
					return
				}
				_go_complete(m, arg...)
			}},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) {
				_c_tags(m, GODOC, "gotags", "-f", nfs.TAGS, "-R", nfs.PWD)
			}},
		}, PlugAction(), LangAction())},
		GODOC: {Name: "godoc", Help: "文档", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				arg[1] = strings.Replace(arg[1], "kit.", "shylinux.com/x/toolkits.", 1)
				arg[1] = strings.Replace(arg[1], "m.", "shylinux.com/x/ice.Message.", 1)
				if m.Cmdy(cli.SYSTEM, GO, "doc", strings.TrimSuffix(arg[1], ".godoc"), kit.Dict(cli.CMD_DIR, arg[2])); m.Append(cli.CMD_ERR) != "" {
					m.Result(m.Append(cli.CMD_OUT))
				}
			}},
		}, PlugAction())},
		MOD: {Name: "mod", Help: "模块", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _mod_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
		SUM: {Name: "sum", Help: "版本", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _sum_show(m, path.Join(arg[2], arg[1])) }},
		}, PlugAction())},
	})
}
