package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _defs_list(m *ice.Message) string {
	return m.OptionDefault(mdb.LIST, ice.Maps{
		"Zone": "zone id auto insert",
		"Hash": "hash auto create",
		"Data": "path auto",
		"Lang": "path auto",
		"Code": "port path auto start order build download",
	}[m.Option(mdb.TYPE)])
}
func _autogen_source(m *ice.Message, main, file string) {
	main = kit.ExtChange(main, SHY)
	m.Cmd(nfs.DEFS, main, `title "{{.Option "name"}}"`+ice.NL)
	m.Cmd(nfs.PUSH, main, ssh.SOURCE+ice.PS+strings.TrimPrefix(file, ice.SRC+ice.PS)+ice.NL)
}
func _autogen_script(m *ice.Message, file string) { m.Cmd(nfs.DEFS, file, _script_template) }
func _autogen_module(m *ice.Message, file string) { m.Cmd(nfs.DEFS, file, _module_template) }
func _autogen_import(m *ice.Message, main string, ctx string, mod string) {
	m.Cmd(nfs.DEFS, main, _main_template)
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
	m.Cmd(nfs.SAVE, main, kit.Join(list, ice.NL))
}
func _autogen_version(m *ice.Message) {
	if mod := _autogen_mod(m, ice.GO_MOD); !nfs.ExistsFile(m, ".git") {
		m.Cmdy(cli.SYSTEM, GIT, ice.INIT)
		m.Cmd(cli.SYSTEM, GIT, "remote", "add", "origin", "https://"+mod)
		m.Cmd(cli.SYSTEM, GIT, "add", ice.GO_MOD, ice.SRC, ice.ETC_MISS_SH)
		m.Cmd(nfs.DEFS, ".gitignore", _git_ignore)
		m.Cmd("web.code.git.repos", mdb.CREATE, nfs.REPOS, "https://"+mod, mdb.NAME, path.Base(mod), nfs.PATH, nfs.PWD)
	}
	m.Cmd(nfs.DEFS, ice.SRC_BINPACK_GO, `package main`+ice.NL)
	m.Cmd(nfs.SAVE, ice.SRC_VERSION_GO, kit.Format(_version_template, _autogen_gits(m, nfs.MODULE, _autogen_mod(m, ice.GO_MOD), tcp.HOSTNAME, ice.Info.Hostname, aaa.USERNAME, ice.Info.Username)))
	m.Cmdy(nfs.DIR, ice.SRC_BINPACK_GO)
	m.Cmdy(nfs.DIR, ice.SRC_VERSION_GO)
	m.Cmdy(nfs.DIR, ice.SRC_MAIN_GO)
}
func _autogen_gits(m *ice.Message, arg ...string) string {
	res := []string{}
	kit.Fetch(_autogen_git(m, arg...), func(k string, v string) {
		res = append(res, kit.Format(`		%s: "%s",`, kit.Capital(k), strings.TrimSpace(v)))
	})
	return kit.Join(res, ice.NL)
}
func _autogen_git(m *ice.Message, arg ...string) ice.Map {
	return kit.Dict(nfs.PATH, kit.Path(""), mdb.TIME, m.Time(), arg,
		mdb.HASH, m.Cmdx(cli.SYSTEM, GIT, "log", "-n1", `--pretty=%H`),
		nfs.REMOTE, m.Cmdx(cli.SYSTEM, GIT, "config", "remote.origin.url"),
		nfs.BRANCH, m.Cmdx(cli.SYSTEM, GIT, "rev-parse", "--abbrev-ref", "HEAD"),
		nfs.VERSION, m.Cmdx(cli.SYSTEM, GIT, "describe", "--tags"),
		web.DOMAIN, kit.Split(m.Option(ice.MSG_USERWEB), "?")[0],
	)
}
func _autogen_mod(m *ice.Message, file string) (mod string) {
	host := web.OptionUserWeb(m).Hostname()
	if host == "" {
		host = path.Base(kit.Path(""))
	} else {
		host = path.Join(host, "x", path.Base(kit.Path("")))
	}
	m.Cmd(nfs.DEFS, file, kit.Format(`module %s

go 1.11
`, host))
	m.Cmd(nfs.CAT, file, func(line string) {
		if strings.HasPrefix(line, "module") {
			mod = kit.Split(line, ice.SP)[1]
		}
	})
	return
}

const AUTOGEN = "autogen"

func init() {
	Index.MergeCommands(ice.Commands{
		AUTOGEN: {Name: "autogen path auto module binpack script relay", Help: "生成", Actions: ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case cli.MAIN:
					m.Cmdy(nfs.DIR, nfs.PWD, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_ROOT, m.Option(nfs.PATH), nfs.DIR_REG, `.*\.go$`)).RenameAppend(nfs.PATH, arg[0])
				}
			}},
			nfs.MODULE: {Name: "module name*=hi help type*=Zone,Hash,Data,Code,Lang main*=main.go@key zone key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.ZONE, m.Option(mdb.NAME), mdb.HELP, m.Option(mdb.NAME))
				m.OptionDefault(mdb.KEY, Prefix(m.Option(mdb.ZONE), m.Option(mdb.NAME)))
				m.Option(mdb.TEXT, kit.Format("`name:\"list %s\" help:\"%s\"`", _defs_list(m), m.Option(mdb.HELP)))
				nfs.OptionFiles(m, nfs.DiskFile)
				if p := path.Join(ice.SRC, m.Option(mdb.ZONE), kit.Keys(m.Option(mdb.NAME), SHY)); !nfs.ExistsFile(m, p) {
					_autogen_source(m, path.Join(ice.SRC, m.Option(cli.MAIN)), p)
					_autogen_script(m, p)
				}
				if p := path.Join(ice.SRC, m.Option(mdb.ZONE), kit.Keys(m.Option(mdb.NAME), GO)); !nfs.ExistsFile(m, p) {
					_autogen_import(m, path.Join(ice.SRC, m.Option(cli.MAIN)), m.Option(mdb.ZONE), _autogen_mod(m, ice.GO_MOD))
					_autogen_module(m, p)
				}
				m.Option(nfs.FILE, path.Join(m.Option(mdb.ZONE), kit.Keys(m.Option(mdb.NAME), GO)))
				_autogen_version(m.Spawn())
			}},
			nfs.SCRIPT: {Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, ice.ETC_MISS_SH, _miss_template)
				defer m.Cmdy(nfs.CAT, ice.ETC_MISS_SH)
				m.Cmdy(nfs.DIR, ice.ETC_MISS_SH)
				m.Cmdy(nfs.DIR, ice.GO_MOD)
				m.Cmdy(nfs.DIR, ice.GO_SUM)
			}},
			BINPACK: {Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(BINPACK, mdb.CREATE); nfs.ExistsFile(m, ice.USR_RELEASE) && m.Option(ice.MSG_USERPOD) == "" {
					const (
						CONF_GO    = "conf.go"
						BINPACK_GO = "binpack.go"
					)
					m.Cmd(nfs.COPY, ice.USR_RELEASE+CONF_GO, ice.USR_ICEBERGS+CONF_GO)
					cli.SystemCmds(m, kit.Format(`cat %s|sed 's/package main/package ice/g' > %s`, ice.SRC_BINPACK_GO, ice.USR_RELEASE+BINPACK_GO))
					m.Cmdy(nfs.DIR, ice.USR_RELEASE+BINPACK_GO)
					m.Cmdy(nfs.DIR, ice.USR_RELEASE+CONF_GO)
				}
				_autogen_version(m)
				m.Cmdy(nfs.CAT, ice.SRC_VERSION_GO)
			}},
			RELAY: {Name: "relay alias username host port init", Help: "跳板", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(COMPILE, RELAY)
				m.Cmdy(nfs.LINK, ice.USR_PUBLISH+m.Option(mdb.ALIAS), ice.USR_PUBLISH+RELAY)
				m.Cmd(nfs.SAVE, path.Join(kit.Env(cli.HOME), ".ssh/"+m.Option(mdb.ALIAS)+".json"),
					kit.Formats(kit.Dict(m.OptionSimple("username,host,port,init"))))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(nfs.CAT, kit.Select(path.Base(ice.SRC_VERSION_GO), arg, 0), kit.Dict(nfs.DIR_ROOT, ice.SRC))
		}},
	})
}

var _miss_template = `#! /bin/sh

if [ -f $PWD/.ish/plug.sh ]; then source $PWD/.ish/plug.sh; elif [ -f $HOME/.ish/plug.sh ]; then source $HOME/.ish/plug.sh; else
	ctx_temp=$(mktemp); if curl -h &>/dev/null; then curl -o $ctx_temp -fsSL https://shylinux.com; else wget -O $ctx_temp -q http://shylinux.com; fi; source $ctx_temp intshell
fi

require miss.sh
ish_miss_prepare_compile
ish_miss_prepare_develop
ish_miss_prepare_operate

ish_miss_make; if [ -n "$*" ]; then ish_miss_serve "$@"; fi
`
var _main_template = `package main

import (
	"shylinux.com/x/ice"
)

func main() { print(ice.Run()) }
`
var _module_template = `package {{.Option "zone"}}

import (
	"shylinux.com/x/ice"
)

type {{.Option "name"}} struct {
	ice.{{.Option "type"}}

	list string {{.Option "text"}}
}

func (s {{.Option "name"}}) List(m *ice.Message, arg ...string) {
	s.{{.Option "type"}}.List(m, arg...)
}

func init() { ice.Cmd("{{.Option "key"}}", {{.Option "name"}}{}) }
`
var _version_template = `package main

import ice "shylinux.com/x/icebergs"

func init() {
	ice.Info.Make = ice.MakeInfo{
%s
	}
}
`
var _script_template = `chapter "{{.Option "name"}}"

field {{.Option "key"}}
`
var _git_ignore = `
src/binpack.go
src/version.go
etc/
bin/
var/
usr/
.*
`
