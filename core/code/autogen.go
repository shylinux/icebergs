package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _defs(m *ice.Message, args ...string) string {
	for i := 0; i < len(args); i += 2 {
		if m.Option(args[i]) == "" {
			m.Option(args[i], args[i+1])
		}
	}
	return m.Option(args[0])
}
func _defs_list(m *ice.Message) string {
	list := []string{mdb.LIST}
	switch m.Option(mdb.TYPE) {
	case "Hash":
		list = append(list, "hash auto create")
	case "Zone":
		list = append(list, "zone id auto insert")
	case "Lists":
		list = append(list, "id auto insert")
	case "Data":
		list = append(list, "path auto")
	case "Code":
		list = append(list, "port path auto start order build download")
	}
	return _defs(m, mdb.LIST, kit.Join(list, ice.SP))
}
func _autogen_module(m *ice.Message, dir string) {
	m.Cmd(nfs.DEFS, dir, `package {{.Option "zone"}}

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
`)
}
func _autogen_import(m *ice.Message, main string, ctx string, mod string) {
	m.Cmd(nfs.DEFS, main, `package main

import (
	"shylinux.com/x/ice"
)

func main() { print(ice.Run()) }
`)

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
func _autogen_script(m *ice.Message, dir string) {
	m.Cmd(nfs.DEFS, dir, `chapter "{{.Option "name"}}"

field "{{.Option "help"}}" {{.Option "key"}}
`)
}
func _autogen_source(m *ice.Message, main, file string) {
	main = strings.ReplaceAll(main, ice.PT+GO, ice.PT+SHY)
	m.Cmd(nfs.DEFS, main, `title "{{.Option "name"}}"
`)
	m.Cmd(nfs.PUSH, main, ice.NL, "source "+strings.TrimPrefix(file, ice.SRC+ice.PS))
}
func _autogen_mod(m *ice.Message, file string) (mod string) {
	host := kit.ParseURLMap(m.Option(ice.MSG_USERWEB))[tcp.HOSTNAME]
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

func _autogen_git(m *ice.Message, arg ...string) ice.Map {
	return kit.Dict("Path", kit.Path(""), "Time", m.Time(), arg,
		"Hash", strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, "log", "-n1", `--pretty=%H`)),
		"Remote", strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, "config", "remote.origin.url")),
		"Branch", strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, "rev-parse", "--abbrev-ref", "HEAD")),
		"Version", strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, "describe", "--tags")),
		"Domain", m.Option(ice.MSG_USERWEB),
	)
}
func _autogen_gits(m *ice.Message, arg ...string) string {
	res := []string{}
	kit.Fetch(_autogen_git(m, arg...), func(k string, v ice.Any) {
		res = append(res, kit.Format(`		%s: "%s",`, k, v))
	})
	return kit.Join(res, ice.NL)
}
func _autogen_version(m *ice.Message) {
	if mod := _autogen_mod(m, ice.GO_MOD); !kit.FileExists(".git") {
		m.Cmdy(cli.SYSTEM, GIT, ice.INIT)
		m.Cmd(cli.SYSTEM, GIT, "remote", "add", "origin", "https://"+mod)
		m.Cmd("web.code.git.repos", mdb.CREATE, "repos", "https://"+mod, mdb.NAME, path.Base(mod), nfs.PATH, nfs.PWD)
		m.Cmd(cli.SYSTEM, GIT, "add", ice.GO_MOD, ice.SRC, ice.ETC_MISS_SH)
		m.Cmd(nfs.DEFS, ".gitignore", kit.Format(`src/binpack.go
src/version.go
etc/
bin/
var/
usr/
.*
`))
	}

	m.Cmd(nfs.DEFS, ice.SRC_BINPACK_GO, kit.Format(`package main
`))

	m.Cmd(nfs.SAVE, ice.SRC_VERSION_GO, kit.Format(`package main

import (
	ice "shylinux.com/x/icebergs"
)

func init() {
	ice.Info.Make = ice.MakeInfo{
%s
	}
}
`, _autogen_gits(m, "Module", _autogen_mod(m, ice.GO_MOD), "HostName", ice.Info.HostName, "UserName", ice.Info.UserName)))

	m.Cmdy(nfs.DIR, ice.SRC_MAIN_GO)
	m.Cmdy(nfs.DIR, ice.SRC_VERSION_GO)
	m.Cmdy(nfs.DIR, ice.SRC_BINPACK_GO)
}

const AUTOGEN = "autogen"

func init() {
	Index.MergeCommands(ice.Commands{
		AUTOGEN: {Name: "autogen path auto create binpack script relay", Help: "生成", Actions: ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case cli.MAIN:
					m.Option(nfs.DIR_ROOT, ice.SRC)
					m.Cmdy(nfs.DIR, nfs.PWD, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_REG, `.*\.go`)).RenameAppend(nfs.PATH, arg[0])
				}
			}},
			mdb.CREATE: {Name: "create name=hi help type=Hash,Zone,Lists,Data,Code main=main.go@key zone key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				_defs(m, mdb.ZONE, m.Option(mdb.NAME), mdb.HELP, m.Option(mdb.NAME))
				_defs(m, mdb.KEY, kit.Keys("web.code", m.Option(mdb.ZONE), m.Option(mdb.NAME)))
				m.Option(mdb.TEXT, kit.Format("`name:\"%s\" help:\"%s\"`", _defs_list(m), m.Option(mdb.HELP)))

				nfs.OptionFiles(m, nfs.DiskFile)
				if p := path.Join(ice.SRC, m.Option(mdb.ZONE), kit.Keys(m.Option(mdb.NAME), GO)); !kit.FileExists(p) {
					_autogen_module(m, p)
					_autogen_import(m, path.Join(ice.SRC, m.Option(cli.MAIN)), m.Option(mdb.ZONE), _autogen_mod(m, ice.GO_MOD))
				}
				if p := path.Join(ice.SRC, m.Option(mdb.ZONE), kit.Keys(m.Option(mdb.NAME), SHY)); !kit.FileExists(p) {
					_autogen_script(m, p)
					_autogen_source(m, path.Join(ice.SRC, m.Option(cli.MAIN)), p)
				}
				m.Option(nfs.FILE, path.Join(m.Option(mdb.ZONE), kit.Keys(m.Option(mdb.NAME), GO)))
				_autogen_version(m.Spawn())
			}},
			ssh.SCRIPT: {Name: "script", Help: "脚本：生成 etc/miss.sh", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, ice.ETC_MISS_SH, m.Conf(web.DREAM, kit.Keym("miss")))
				defer m.Cmdy(nfs.CAT, ice.ETC_MISS_SH)

				m.Cmdy(nfs.DIR, ice.ETC_MISS_SH)
				m.Cmdy(nfs.DIR, ice.GO_MOD)
				m.Cmdy(nfs.DIR, ice.GO_SUM)
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.TRASH, path.Join(ice.SRC, m.Option(nfs.PATH)))
			}},
			BINPACK: {Name: "binpack", Help: "打包：生成 src/binpack.go", Hand: func(m *ice.Message, arg ...string) {
				_autogen_version(m)
				if m.Cmd(BINPACK, mdb.CREATE); kit.FileExists(ice.USR_RELEASE) && m.Option(ice.MSG_USERPOD) == "" {
					m.Cmd(nfs.COPY, path.Join(ice.USR_RELEASE, "conf.go"), path.Join(ice.USR_ICEBERGS, "conf.go"))
					m.Cmd(cli.SYSTEM, "sh", "-c", `cat src/binpack.go|sed 's/package main/package ice/g' > usr/release/binpack.go`)
					m.Cmdy(nfs.DIR, "usr/release/binpack.go")
					m.Cmdy(nfs.DIR, "usr/release/conf.go")
				}
				m.Cmdy(nfs.CAT, "src/version.go")
			}},
			RELAY: {Name: "relay alias username host port list", Help: "跳板", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(COMPILE, RELAY)
				m.Cmd(nfs.LINK, path.Join(ice.USR_PUBLISH, m.Option(mdb.ALIAS)), RELAY)
				m.Cmd(nfs.SAVE, path.Join(kit.Env(cli.HOME), ".ssh/"+m.Option(mdb.ALIAS)+".json"),
					kit.Formats(kit.Dict(m.OptionSimple("username,host,port,list"))))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(nfs.CAT, kit.Select("version.go", arg, 0), kit.Dict(nfs.DIR_ROOT, ice.SRC))
		}},
	})
}
