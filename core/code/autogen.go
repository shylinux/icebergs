package code

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _defs(m *ice.Message, args ...string) {
	for i := 0; i < len(args); i += 2 {
		if m.Option(args[i]) == "" {
			m.Option(args[i], args[i+1])
		}
	}
}
func _autogen_script(m *ice.Message, dir string) {
	buf, _ := kit.Render(`chapter "{{.Option "name"}}"
field "{{.Option "help"}}" {{.Option "key"}}
`, m)
	m.Cmd(nfs.DEFS, dir, string(buf))
}
func _autogen_source(m *ice.Message, zone, name string) {
	m.Cmd(nfs.PUSH, ice.SRC_MAIN_SHY, "\n", kit.SSH_SOURCE+` `+path.Join(zone, kit.Keys(name, SHY)), "\n")
}
func _autogen_module(m *ice.Message, dir string) {
	buf, _ := kit.Render(`package {{.Option "zone"}}

import (
	"shylinux.com/x/ice"
)

type {{.Option "name"}} struct {
	ice.{{.Option "type"}}
}

func (h {{.Option "name"}}) List(m *ice.Message, arg ...string) {
	h.{{.Option "type"}}.List(m, arg...)
}

func init() {
	ice.Cmd("{{.Option "key"}}", &{{.Option "name"}}{}, []*ice.Show{
		{Name: "{{.Option "list"}}", Help: "{{.Option "help"}}"},
	})
}
`, m)
	m.Cmd(nfs.SAVE, dir, string(buf))
}
func _autogen_import(m *ice.Message, main string, ctx string, mod string) (list []string) {
	m.Cmd(nfs.DEFS, ice.GO_MOD, kit.Format(`module %s

go 1.11
`, path.Base(kit.Path(""))))

	m.Cmd(nfs.DEFS, main, `package main

import "shylinux.com/x/ice"

func main() { print(ice.Run()) }
`)

	done := false
	m.Cmd(nfs.CAT, main, func(line string, index int) {
		if list = append(list, line); done {
			return
		}
		if strings.HasPrefix(line, "import (") {
			list = append(list, kit.Format(`	_ "%s/src/%s"`, mod, ctx), "")
			done = true
		} else if strings.HasPrefix(line, "import") {
			list = append(list, "", kit.Format(`import _ "%s/src/%s"`, mod, ctx), "")
			done = true
		}
	})

	m.Cmd(nfs.SAVE, main, strings.Join(list, ice.NL))
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
	if _, e := os.Stat(".git"); os.IsNotExist(e) {
		m.Cmdy(cli.SYSTEM, "git", "init")
	}
	if _, e := os.Stat("go.mod"); os.IsNotExist(e) {
		m.Cmdy(cli.SYSTEM, "go", "mod", "init", path.Base(kit.Path("")))
	}

	file := ice.SRC_VERSION
	m.Cmd(nfs.SAVE, file, kit.Format(`package main

import (
	"shylinux.com/x/icebergs"
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
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		AUTOGEN: {Name: "autogen path auto create binpack script", Help: "生成", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create main=src/main.go@key key= zone= type=Zone,Hash,List,Data name=hi list= help=", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				_defs(m, "zone", m.Option("name"), "help", m.Option("name"))
				_defs(m, "key", kit.Keys("web.code", m.Option("zone"), m.Option("name")))
				switch m.Option("type") {
				case "Zone":
					_defs(m, "list", "list zone id auto insert")
				case "Hash":
					_defs(m, "list", "list hash auto create")
				case "List":
					_defs(m, "list", "list id auto insert")
				case "Data":
					_defs(m, "list", "list path auto upload")
				}

				if p := path.Join(kit.SSH_SRC, m.Option(kit.MDB_ZONE), kit.Keys(m.Option(kit.MDB_NAME), SHY)); !kit.FileExists(p) {
					_autogen_script(m, p)
					_autogen_source(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_NAME))
				}

				if p := path.Join(kit.SSH_SRC, m.Option(kit.MDB_ZONE), kit.Keys(m.Option(kit.MDB_NAME), GO)); !kit.FileExists(p) {
					_autogen_module(m, p)
					_autogen_import(m, m.Option(kit.MDB_MAIN), m.Option(kit.MDB_ZONE), _autogen_mod(m, ice.GO_MOD))
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_MAIN:
					m.Option(nfs.DIR_REG, `.*\.go`)
					m.Cmdy(nfs.DIR, kit.SSH_SRC, "path,size,time")
					m.RenameAppend(kit.MDB_PATH, arg[0])
				}
			}},
			BINPACK: {Name: "binpack", Help: "打包：生成 src/binpack.go", Hand: func(m *ice.Message, arg ...string) {
				_autogen_version(m)
				m.Cmd(BINPACK, mdb.CREATE)
			}},
			"script": {Name: "script", Help: "脚本：生成 etc/miss.sh", Hand: func(m *ice.Message, arg ...string) {
				_autogen_miss(m)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option(nfs.DIR_ROOT, kit.SSH_SRC); len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				m.Cmdy(nfs.DIR, kit.Select("./", arg, 0))
			} else {
				m.Cmdy(nfs.CAT, arg[0])
			}
		}},
	}})
}
