package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const SEARCHS = "searchs"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SEARCHS: {Name: "searchs", Help: "搜索", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			"/search": {Name: "/search", Help: "搜索引擎", Action: map[string]*ice.Action{
				mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(m.Space(m.Option(POD)), mdb.RENDER, arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if kit.Contains(arg[1], ";") {
					arg = kit.Split(arg[1], ";", ";", ";")
				}

				if m.Cmdy(m.Space(m.Option(POD)), mdb.SEARCH, arg); arg[1] == "" {
					return
				}
				m.Cmd(mdb.INSERT, m.Prefix(SEARCHS), "", mdb.HASH,
					"name", arg[1], "type", arg[0], "text", kit.Select("", arg, 2))
			}},
		}})
}
