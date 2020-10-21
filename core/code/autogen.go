package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

func _autogen_script(m *ice.Message, dir string) {
	if b, e := kit.Render(m.Conf(AUTOGEN, "meta.shy"), m); m.Assert(e) {
		m.Cmd(nfs.SAVE, dir, string(b))
	}
}
func _autogen_source(m *ice.Message, name string) {
	m.Cmd("nfs.file", "append", "src/main.shy", "\n", `source `+name+"/"+name+".shy", "\n")
}
func _autogen_index(m *ice.Message, dir string, from string, ctx string) {
	list := []string{}

	up, low := "", ""
	key := strings.ToUpper(ctx)
	m.Option(nfs.CAT_CB, func(line string, index int) {
		if strings.HasPrefix(line, "package") {
			line = "package " + ctx
		}
		if up == "" && strings.HasPrefix(line, "const") {
			if ls := kit.Split(line); len(ls) > 3 {
				up, low = ls[1], ls[3]
			}
		}
		if up != "" {
			line = strings.ReplaceAll(line, up, key)
			line = strings.ReplaceAll(line, low, ctx)
		}

		list = append(list, line)
	})
	m.Cmd(nfs.CAT, from)

	m.Cmdy(nfs.SAVE, dir, strings.Join(list, "\n"))
}
func _autogen_main(m *ice.Message, file string, mod string, ctx string) {
	list := []string{}

	m.Option(nfs.CAT_CB, func(line string, index int) {
		list = append(list, line)
		if strings.HasPrefix(line, "import (") {
			list = append(list, kit.Format(`	_ "%s/src/%s"`, mod, ctx), "")
		}
	})
	m.Cmd(nfs.CAT, file)

	m.Cmd(nfs.SAVE, file, strings.Join(list, "\n"))
}
func _autogen_mod(m *ice.Message, file string) (mod string) {
	m.Option(nfs.CAT_CB, func(line string, index int) {
		if strings.HasPrefix(line, "module") {
			mod = strings.Split(line, " ")[1]
			m.Info("module", mod)
		}
	})
	m.Cmd(nfs.CAT, "go.mod")
	return
}

const AUTOGEN = "autogen"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			AUTOGEN: {Name: AUTOGEN, Help: "生成", Value: kit.Data(
				kit.MDB_FIELD, "time,id,name,from", "shy", `
chapter "{{.Option "name"}}"
field "{{.Option "name"}}" web.code.{{.Option "name"}}.{{.Option "name"}}
`,
			)},
		},
		Commands: map[string]*ice.Command{
			AUTOGEN: {Name: "autogen path auto create", Help: "生成", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create main=src/main.go@key name=hi@key from=usr/icebergs/misc/zsh/zsh.go@key", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					if p := path.Join("src", m.Option("name"), m.Option("name")+".shy"); !kit.FileExists(p) {
						_autogen_script(m, p)
						_autogen_source(m, m.Option("name"))
					}

					if p := path.Join("src", m.Option("name"), m.Option("name")+".go"); !kit.FileExists(p) {
						_autogen_index(m, p, m.Option("from"), m.Option("name"))
						_autogen_main(m, m.Option("main"), _autogen_mod(m, "go.mod"), m.Option("name"))
					}
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "main":
						m.Cmdy(nfs.DIR, "src", "path,size,time")
						m.Sort(kit.MDB_PATH)
					case "from":
						m.Option(nfs.DIR_DEEP, true)
						m.Cmdy(nfs.DIR, "usr/icebergs", "path,size,time")
						m.Sort(kit.MDB_PATH)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(nfs.DIR_ROOT, "src"); len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
					m.Cmdy(nfs.DIR, kit.Select("./", arg, 0))
				} else {
					m.Cmdy(nfs.CAT, arg[0])
				}
			}},
		},
	}, nil)
}
