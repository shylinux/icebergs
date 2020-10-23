package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const SPIDE = "spide"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			SPIDE: {Name: "spide path file auto", Help: "结构图", Meta: kit.Dict(
				"display", "/plugin/story/spide.js",
			), Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "ctx.command"},
				code.INNER:  {Name: "web.code.inner"},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 仓库列表
					m.Option(ice.MSG_DISPLAY, "table")
					m.Cmdy(TOTAL, arg)
					return
				}
				if len(arg) == 1 {
					// 目录列表
					m.Option(nfs.DIR_ROOT, arg[0])
					m.Option(nfs.DIR_DEEP, "true")
					m.Cmdy(nfs.DIR, "./")
					return
				}

				if m.Option(cli.CMD_DIR, arg[0]); strings.HasSuffix(arg[1], ".go") {
					tags := m.Cmdx(cli.SYSTEM, "gotags", arg[1])

					for _, line := range strings.Split(tags, "\n") {
						if len(line) == 0 || strings.HasPrefix(line, "!_") {
							continue
						}

						ls := kit.Split(line, "\t ", "\t ", "\t ")
						name := ls[3] + ":" + ls[0]
						switch ls[3] {
						case "m":
							if strings.HasPrefix(ls[5], "ctype") {
								name = strings.TrimPrefix(ls[5], "ctype:") + ":" + ls[0]
							} else if strings.HasPrefix(ls[6], "ntype") {
								name = "-" + ls[0]
							} else {

							}
						case "w":
							t := ls[len(ls)-1]
							name = "-" + ls[0] + ":" + strings.TrimPrefix(t, "type:")
						}

						m.Push("name", name)
						m.Push("file", ls[1])
						m.Push("line", strings.TrimSuffix(ls[2], ";\""))
						m.Push("type", ls[3])
						m.Push("extra", strings.Join(ls[4:], " "))
					}
				} else {
					tags := m.Cmdx(cli.SYSTEM, "ctags", "-f", "-", arg[1])

					for _, line := range strings.Split(tags, "\n") {
						if len(line) == 0 || strings.HasPrefix(line, "!_") {
							continue
						}

						ls := kit.Split(line, "\t ", "\t ", "\t ")
						m.Push("name", ls[0])
						m.Push("file", ls[1])
						m.Push("line", "1")
					}
				}
				m.SortInt(kit.MDB_LINE)
			}},
		},
	}, nil)
}
