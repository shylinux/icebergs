package vim

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const INPUT = "input"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			INPUT: {Name: INPUT, Help: "输入法", Value: kit.Data(
				kit.MDB_FIELD, "time,id,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			INPUT: {Name: "sync id=auto auto export import", Help: "同步流", Action: map[string]*ice.Action{
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(INPUT), "", mdb.LIST)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(INPUT), "", mdb.LIST)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select(m.Conf(INPUT, kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, m.Prefix(INPUT), "", mdb.LIST, kit.MDB_ID, arg)
			}},

			"/input": {Name: "/input", Help: "补全", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if arg[0] = strings.TrimSpace(arg[0]); strings.HasPrefix(arg[0], "ice ") {
					list := kit.Split(arg[0])
					switch list[1] {
					case "add":
						// ice add 想你 shwq [9999 [person]]
						m.Cmd("web.code.input.wubi", "insert", "text", list[2], "code", list[3],
							"weight", kit.Select("999999", list, 4), "zone", kit.Select("person", list, 5))
						arg[0] = list[3]
					default:
						// ice command
						if m.Cmdy(list[1:]); strings.TrimSpace(m.Result()) == "" {
							m.Table()
						}
						if strings.TrimSpace(m.Result()) == "" {
							m.Cmdy(cli.SYSTEM, list[1:])
						}
						m.Cmd(mdb.INSERT, m.Prefix(INPUT), "", mdb.LIST, kit.MDB_TYPE, "cmd",
							kit.MDB_NAME, strings.TrimSpace(strings.Join(list[1:], " ")), kit.MDB_TEXT, m.Result())
						m.Echo("%s\n", arg[0])
						return
					}
				}

				// 词汇列表
				m.Cmd("web.code.input.wubi", "word", arg[0]).Table(func(index int, value map[string]string, head []string) {
					m.Echo("%s\n", value["text"])
				})

				m.Cmd(mdb.INSERT, m.Prefix(INPUT), "", mdb.LIST, kit.MDB_TYPE, "wubi", kit.MDB_NAME, arg[0], kit.MDB_TEXT, m.Result())
			}},
		},
	})
}
