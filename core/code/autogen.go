package code

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

const AUTOGEN = "autogen"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			AUTOGEN: {Name: AUTOGEN, Help: "生成器", Value: kit.Data(
				kit.MDB_FIELD, "time,id,name,from",
			)},
		},
		Commands: map[string]*ice.Command{
			AUTOGEN: {Name: "autogen auto 创建", Help: "生成器", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name=hi@key from=usr/icebergs/misc/zsh/zsh.go@key", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					if p := path.Join("src", m.Option("name"), m.Option("name")+".shy"); !kit.FileExists(p) {
						m.Cmd(nfs.SAVE, p, `chapter "`+m.Option("name")+`"`, "\n", `field "`+m.Option("name")+`" web.code.`+m.Option("name")+"."+m.Option("name"))
						m.Cmd("nfs.file", "append", "src/main.shy", "\n", `source `+m.Option("name")+"/"+m.Option("name")+".shy", "\n")
					}

					p := path.Join("src", m.Option("name"), m.Option("name")+".go")
					if kit.FileExists(p) {
						return
					}

					// module file
					list := []string{}
					up, low := "", ""
					name := m.Option("name")
					key := strings.ToUpper(m.Option("name"))
					m.Option(nfs.CAT_CB, func(line string, index int) {
						if strings.HasPrefix(line, "package") {
							line = "package " + m.Option("name")
						}
						if up == "" && strings.HasPrefix(line, "const") {
							if ls := kit.Split(line); len(ls) > 3 {
								up, low = ls[1], ls[3]
							}
						}
						if up != "" {
							line = strings.ReplaceAll(line, up, key)
							line = strings.ReplaceAll(line, low, name)
						}

						list = append(list, line)
					})
					m.Cmd(nfs.CAT, m.Option("from"))
					m.Cmd(nfs.SAVE, p, strings.Join(list, "\n"))

					// go.mod
					mod := ""
					m.Option(nfs.CAT_CB, func(line string, index int) {
						if strings.HasPrefix(line, "module") {
							mod = strings.Split(line, " ")[1]
							m.Info("module", mod)
						}
					})
					m.Cmd(nfs.CAT, "go.mod")

					// src/main.go
					list = []string{}
					m.Option(nfs.CAT_CB, func(line string, index int) {
						list = append(list, line)
						if strings.HasPrefix(line, "import (") {
							list = append(list, `	_ "`+mod+"/src/"+m.Option("name")+`"`, "")
						}
					})
					m.Cmd(nfs.CAT, "src/main.go")
					m.Cmd(nfs.SAVE, "src/main.go", strings.Join(list, "\n"))

					m.Cmdy(mdb.INSERT, m.Prefix(AUTOGEN), "", mdb.LIST, arg)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "from":
						m.Option(nfs.DIR_DEEP, true)
						m.Cmdy(nfs.DIR, "usr/icebergs")
						m.Sort("path")
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select(m.Conf(m.Prefix(AUTOGEN), kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, m.Prefix(AUTOGEN), "", mdb.LIST, kit.MDB_ID, arg)
			}},
		},
	}, nil)
}
