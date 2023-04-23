package code

import (
	"bytes"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _autogen_list(m *ice.Message) string {
	return m.OptionDefault(mdb.LIST, ice.Maps{
		"Zone": "zone id auto insert",
		"Hash": "hash auto create",
		"Data": "path auto",
		"Lang": "path auto",
		"Code": "port path auto start order build download",
	}[m.Option(mdb.TYPE)])
}
func _autogen_source(m *ice.Message, main, file string) {
	m.Cmd(nfs.DEFS, main, nfs.Template(m, ice.SRC_MAIN_SHY))
	m.Cmd(nfs.PUSH, main, ssh.SOURCE+lex.SP+strings.TrimPrefix(file, ice.SRC+nfs.PS)+lex.NL)
}
func _autogen_script(m *ice.Message, file string) { m.Cmd(nfs.DEFS, file, nfs.Template(m, "demo.shy")) }
func _autogen_module(m *ice.Message, file string) { m.Cmd(nfs.DEFS, file, nfs.Template(m, "demo.go")) }
func _autogen_import(m *ice.Message, main string, ctx string, mod string) {
	m.Cmd(nfs.DEFS, main, nfs.Template(m, ice.SRC_MAIN_GO))
	begin, done, list := false, false, []string{}
	m.Cmd(nfs.CAT, main, func(line string, index int) {
		if begin && !done && strings.HasPrefix(line, ")") {
			done, list = true, append(list, "", kit.Format(`	_ "%s/src/%s"`, mod, ctx))
		}
		if list = append(list, line); done {
			return
		}
		if strings.HasPrefix(line, "import (") {
			begin = true
		} else if strings.HasPrefix(line, "import") {
			done, list = true, append(list, kit.Format(`import _ "%s/src/%s"`, mod, ctx))
		}
	})
	m.Cmd(nfs.SAVE, main, kit.Join(list, lex.NL))
	m.Cmd(cli.SYSTEM, "goimports", "-w", main)
}
func _autogen_version(m *ice.Message) string {
	if mod := _autogen_mod(m, ice.GO_MOD); !nfs.Exists(m, ".git") {
		m.Cmd(REPOS, "init", nfs.ORIGIN, kit.Select(m.Option(ice.MSG_USERHOST), ice.Info.Make.Domain)+"/x/"+path.Base(mod), mdb.NAME, path.Base(mod), nfs.PATH, nfs.PWD)
		defer m.Cmd(REPOS, "add", kit.Dict(nfs.REPOS, path.Base(mod), nfs.FILE, nfs.SRC))
		defer m.Cmd(REPOS, "add", kit.Dict(nfs.REPOS, path.Base(mod), nfs.FILE, "go.mod"))
	}
	m.Cmd(nfs.DEFS, ".gitignore", nfs.Template(m, "gitignore"))
	m.Cmd(nfs.DEFS, ice.SRC_BINPACK_GO, nfs.Template(m, ice.SRC_BINPACK_GO))
	m.Cmd(nfs.SAVE, ice.SRC_VERSION_GO, kit.Format(nfs.Template(m, ice.SRC_VERSION_GO), _autogen_gits(m, nfs.MODULE, _autogen_mod(m, ice.GO_MOD), tcp.HOSTNAME, ice.Info.Hostname)))
	m.Cmd(cli.SYSTEM, "gofmt", "-w", ice.SRC_VERSION_GO)
	m.Cmdy(nfs.DIR, ice.SRC_BINPACK_GO)
	m.Cmdy(nfs.DIR, ice.SRC_VERSION_GO)
	m.Cmdy(nfs.DIR, ice.SRC_MAIN_GO)
	return ice.SRC_VERSION_GO
}
func _autogen_gits(m *ice.Message, arg ...string) string {
	res := []string{}
	kit.For(_autogen_git(m, arg...), func(k, v string) {
		res = append(res, kit.Format(`		%s: "%s",`, kit.Capital(k), strings.TrimSpace(v)))
	})
	return kit.Join(res, lex.NL)
}
func _autogen_git(m *ice.Message, arg ...string) ice.Map {
	return kit.Dict(arg,
		mdb.TIME, m.Time(), nfs.PATH, kit.Path(""), web.DOMAIN, web.UserHost(m),
		mdb.HASH, m.Cmdx(cli.SYSTEM, GIT, "log", "-n1", `--pretty=%H`),
		nfs.REMOTE, m.Cmdx(cli.SYSTEM, GIT, "config", "remote.origin.url"),
		nfs.BRANCH, m.Cmdx(cli.SYSTEM, GIT, "rev-parse", "--abbrev-ref", "HEAD"),
		nfs.VERSION, m.Cmdx(cli.SYSTEM, GIT, "describe", "--tags"),
		aaa.EMAIL, m.Cmdx(cli.SYSTEM, GIT, "config", "user.email"),
		aaa.USERNAME, kit.Select(ice.Info.Username, m.Cmdx(cli.SYSTEM, GIT, "config", "user.name")),
	)
}
func _autogen_mod(m *ice.Message, file string) (mod string) {
	host := kit.ParseURL(kit.Select(m.Option(ice.MSG_USERHOST), ice.Info.Make.Domain)).Hostname()
	if host == "" {
		host = path.Base(kit.Path(""))
	} else {
		host = path.Join(host, "x", path.Base(kit.Path("")))
	}
	m.Cmd(nfs.DEFS, file, nfs.Template(m, ice.GO_MOD, host))
	m.Cmd(nfs.CAT, file, func(line string) {
		kit.If(strings.HasPrefix(line, nfs.MODULE), func() { mod = kit.Split(line, lex.SP)[1] })
	})
	return
}

const AUTOGEN = "autogen"

func init() {
	const (
		VERSION = "version"
	)
	Index.MergeCommands(ice.Commands{
		AUTOGEN: {Name: "autogen path auto script module version", Help: "生成", Actions: ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case cli.MAIN:
					m.Cmdy(nfs.DIR, nfs.PWD, nfs.PATH, kit.Dict(nfs.DIR_ROOT, ice.SRC, nfs.DIR_REG, kit.ExtReg(GO)))
				case mdb.ZONE, mdb.NAME:
					m.Cmdy(nfs.DIR, nfs.PWD, mdb.NAME, kit.Dict(nfs.DIR_ROOT, ice.SRC, nfs.DIR_TYPE, nfs.DIR))
				case mdb.KEY:
					m.Push(arg[0], Prefix(m.Option(mdb.ZONE), m.Option(mdb.NAME)))
				}
			}},
			nfs.SCRIPT: {Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, ice.ETC_MISS_SH, nfs.Template(m, ice.ETC_MISS_SH))
				m.Cmdy(nfs.DIR, ice.ETC_MISS_SH).Cmdy(nfs.CAT, ice.ETC_MISS_SH)
			}},
			nfs.MODULE: {Name: "module name*=hi help type*=Zone,Hash,Data,Code,Lang main*=main.go@key zone key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.ZONE, m.Option(mdb.NAME), mdb.HELP, m.Option(mdb.NAME))
				m.OptionDefault(mdb.KEY, Prefix(m.Option(mdb.ZONE), m.Option(mdb.NAME)))
				m.Option(nfs.FILE, path.Join(m.Option(mdb.ZONE), kit.Keys(m.Option(mdb.NAME), GO)))
				m.Option(mdb.TEXT, kit.Format("`name:\"list %s\" help:\"%s\"`", _autogen_list(m), m.Option(mdb.HELP)))
				defer _autogen_version(m.Spawn())
				if p := path.Join(ice.SRC, kit.ExtChange(m.Option(nfs.FILE), SHY)); !nfs.Exists(m, p) {
					_autogen_source(m, kit.ExtChange(path.Join(ice.SRC, m.Option(cli.MAIN)), SHY), p)
					_autogen_script(m, p)
				}
				if p := path.Join(ice.SRC, m.Option(nfs.FILE)); !nfs.Exists(m, p) {
					_autogen_import(m, path.Join(ice.SRC, m.Option(cli.MAIN)), m.Option(mdb.ZONE), _autogen_mod(m, ice.GO_MOD))
					_autogen_module(m, p)
				}
			}},
			DEVPACK: {Help: "开发", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(WEBPACK, mdb.REMOVE) }},
			WEBPACK: {Help: "打包", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(WEBPACK, mdb.CREATE) }},
			BINPACK: {Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				const (
					USR_RELEASE_CONF_GO    = "usr/release/conf.go"
					USR_RELEASE_BINPACK_GO = "usr/release/binpack.go"
				)
				if m.Cmd(BINPACK, mdb.CREATE); nfs.Exists(m, ice.USR_RELEASE) && m.Option(ice.MSG_USERPOD) == "" {
					nfs.CopyFile(m, USR_RELEASE_BINPACK_GO, ice.SRC_BINPACK_GO, func(buf []byte, offset int) []byte {
						kit.If(offset == 0, func() { buf = bytes.Replace(buf, []byte("package main"), []byte("package ice"), 1) })
						return buf
					})
					m.Cmd(nfs.COPY, USR_RELEASE_CONF_GO, ice.USR_ICEBERGS+"conf.go")
					m.Cmdy(nfs.DIR, USR_RELEASE_BINPACK_GO)
					m.Cmdy(nfs.DIR, USR_RELEASE_CONF_GO)
				}
				m.Cmdy(nfs.CAT, _autogen_version(m))
			}},
			VERSION: {Help: "版本", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, _autogen_version(m)) }},
		}},
	})
}
