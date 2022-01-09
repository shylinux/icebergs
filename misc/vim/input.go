package vim

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const INPUT = "input"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		INPUT: {Name: INPUT, Help: "输入法", Value: kit.Data(
			mdb.FIELD, "time,id,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		"/input": {Name: "/input", Help: "输入法", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Cmdy(TAGS, arg[0]); len(m.Result()) > 0 {
				return // 代码补全
			}

			if arg[0] = strings.TrimSpace(arg[0]); strings.HasPrefix(arg[0], "ice ") {
				switch list := kit.Split(arg[0]); list[1] {
				case "add": // ice add 想你 shwq [person [9999]]
					m.Cmd("web.code.input.wubi", mdb.INSERT, mdb.TEXT, list[2], cli.CODE, list[3],
						mdb.ZONE, kit.Select("person", list, 4), mdb.VALUE, kit.Select("999999", list, 5),
					)
					arg[0] = list[3]
				default: // 执行命令
					if m.Cmdy(list[1:]); strings.TrimSpace(m.Result()) == "" {
						m.Table()
					}
					if strings.TrimSpace(m.Result()) == "" {
						m.Cmdy(cli.SYSTEM, list[1:])
					}
					m.Cmd(INPUT, mdb.INSERT, mdb.TYPE, "cmd", mdb.NAME, strings.TrimSpace(strings.Join(list[1:], ice.SP)), mdb.TEXT, m.Result())
					m.Echo("%s\n", arg[0])
					return
				}
			}

			// 词汇列表
			m.Option(ice.CACHE_LIMIT, "10")
			m.Cmd("web.code.input.wubi", "word", arg[0]).Table(func(index int, value map[string]string, head []string) {
				m.Echo("%s\n", value[mdb.TEXT])
			})
			m.Cmd(INPUT, mdb.INSERT, mdb.TYPE, "wubi", mdb.NAME, arg[0], mdb.TEXT, m.Result())
		}},
		INPUT: {Name: "input id auto export import", Help: "输入法", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {},
		}, mdb.ListAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.ListSelect(m, arg...)
		}},
	}})
}
