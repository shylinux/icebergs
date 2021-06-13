package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

func _autogen_script(m *ice.Message, dir string) {
	if b, e := kit.Render(m.Conf(AUTOGEN, kit.Keym(SHY)), m); m.Assert(e) {
		m.Cmd(nfs.DEFS, dir, string(b))
	}
}
func _autogen_source(m *ice.Message, name string) {
	m.Cmd(nfs.PUSH, ice.SRC_MAIN, "\n", kit.SSH_SOURCE+` `+path.Join(name, kit.Keys(name, SHY)), "\n")
}
func _autogen_module(m *ice.Message, dir string, ctx string, from string) (list []string) {
	name, value := "", ""
	key := strings.ToUpper(ctx)
	m.Cmd(nfs.CAT, from, func(line string, index int) {
		if strings.HasPrefix(line, "package") {
			line = "package " + ctx
		}

		if name == "" && strings.HasPrefix(line, "const") {
			if ls := kit.Split(line); len(ls) > 3 {
				name, value = ls[1], ls[3]
			}
		}
		if name != "" {
			line = strings.ReplaceAll(line, name, key)
			line = strings.ReplaceAll(line, value, ctx)
		}

		list = append(list, line)
	})

	m.Cmd(nfs.SAVE, dir, strings.Join(list, "\n"))
	return
}
func _autogen_import(m *ice.Message, main string, ctx string, mod string) (list []string) {
	m.Cmd(nfs.CAT, main, func(line string, index int) {
		if list = append(list, line); strings.HasPrefix(line, "import (") {
			list = append(list, kit.Format(`	_ "%s/src/%s"`, mod, ctx), "")
		}
	})

	m.Cmd(nfs.SAVE, main, strings.Join(list, "\n"))
	return
}
func _autogen_mod(m *ice.Message, file string) (mod string) {
	m.Cmd(nfs.CAT, file, func(line string, index int) {
		if strings.HasPrefix(line, "module") {
			mod = strings.Split(line, " ")[1]
			m.Logs("module", mod)
		}
	})
	return
}

func _autogen_version(m *ice.Message) {
	file := ice.SRC_VERSION
	m.Cmd(nfs.SAVE, file, kit.Format(`package main

import (
	"github.com/shylinux/icebergs"
)

func init() {
	ice.Info.Make.Time = "%s"
	ice.Info.Make.Hash = "%s"
	ice.Info.Make.Remote = "%s"
	ice.Info.Make.Branch = "%s"
	ice.Info.Make.Version = "%s"
	ice.Info.Make.HostName = "%s"
	ice.Info.Make.UserName = "%s"
}
`,
		m.Time(),
		strings.TrimSpace(m.Cmdx(cli.SYSTEM, "git", "log", "-n1", `--pretty=%H`)),
		strings.TrimSpace(m.Cmdx(cli.SYSTEM, "git", "config", "remote.origin.url")),
		strings.TrimSpace(m.Cmdx(cli.SYSTEM, "git", "rev-parse", "--abbrev-ref", "HEAD")),
		strings.TrimSpace(m.Cmdx(cli.SYSTEM, "git", "describe", "--tags")),
		ice.Info.HostName, ice.Info.UserName,
	))
	defer m.Cmdy(nfs.CAT, file)

	m.Cmdy(nfs.DIR, file, "time,size,line,path")
	m.Cmdy(nfs.DIR, ice.SRC_BINPACK, "time,size,line,path")
	m.Cmdy(nfs.DIR, ice.SRC_MAIN_GO, "time,size,line,path")
}
func _autogen_miss(m *ice.Message) {
	m.Cmd(nfs.DEFS, ice.ETC_MISS, m.Conf(web.DREAM, kit.Keym("miss")))
	defer m.Cmdy(nfs.CAT, ice.ETC_MISS)

	m.Cmdy(nfs.DIR, ice.ETC_MISS, "time,size,line,path")
	m.Cmdy(nfs.DIR, ice.GO_MOD, "time,size,line,path")
	m.Cmdy(nfs.DIR, ice.GO_SUM, "time,size,line,path")
}

const AUTOGEN = "autogen"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			AUTOGEN: {Name: "autogen path auto create binpack script", Help: "生成", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create main=src/main.go@key name=hi@key from=usr/icebergs/misc/bash/bash.go@key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
					if p := path.Join(kit.SSH_SRC, m.Option(kit.MDB_NAME), kit.Keys(m.Option(kit.MDB_NAME), SHY)); !kit.FileExists(p) {
						_autogen_script(m, p)
						_autogen_source(m, m.Option(kit.MDB_NAME))
					}

					if p := path.Join(kit.SSH_SRC, m.Option(kit.MDB_NAME), kit.Keys(m.Option(kit.MDB_NAME), GO)); !kit.FileExists(p) {
						_autogen_module(m, p, m.Option(kit.MDB_NAME), m.Option(kit.MDB_FROM))
						_autogen_import(m, m.Option(kit.MDB_MAIN), m.Option(kit.MDB_NAME), _autogen_mod(m, "go.mod"))
					}
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_MAIN:
						m.Option(nfs.DIR_REG, `.*\.go`)
						m.Cmdy(nfs.DIR, kit.SSH_SRC, "path,size,time")
						m.RenameAppend(kit.MDB_PATH, arg[0])

					case kit.MDB_FROM:
						m.Option(nfs.DIR_DEEP, true)
						m.Option(nfs.DIR_REG, `.*\.go`)
						m.Cmdy(nfs.DIR, kit.SSH_SRC, "path,size,time")
						m.Cmdy(nfs.DIR, "usr/icebergs/misc/", "path,size,time")
						m.RenameAppend(kit.MDB_PATH, arg[0])
					}
				}},
				BINPACK: {Name: "binpack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					_autogen_version(m)
					m.Cmd(BINPACK, mdb.CREATE)
				}},
				"script": {Name: "script", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
					_autogen_miss(m)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(nfs.DIR_ROOT, kit.SSH_SRC); len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
					m.Cmdy(nfs.DIR, kit.Select("./", arg, 0))
				} else {
					m.Cmdy(nfs.CAT, arg[0])
				}
			}},
		},
		Configs: map[string]*ice.Config{
			AUTOGEN: {Name: AUTOGEN, Help: "生成", Value: kit.Data(
				SHY, `chapter "{{.Option "name"}}"
field "{{.Option "name"}}" web.code.{{.Option "name"}}.{{.Option "name"}}
`,
			)},
		},
	})
}
