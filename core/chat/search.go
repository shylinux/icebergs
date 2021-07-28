package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const P_SEARCH = "/search"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			P_SEARCH: {Name: P_SEARCH, Help: "搜索", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, P_SEARCH, m.Prefix(P_SEARCH))
			}},
			P_SEARCH: {Name: P_SEARCH, Help: "搜索引擎", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(P_SEARCH, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						if value = kit.GetMeta(value); arg[1] != "" && !kit.Contains(value[kit.MDB_NAME], arg[1]) {
							return
						}
						m.PushSearch(kit.SSH_CMD, P_SEARCH, value)
					})
				}},
				mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(m.Space(m.Option(cli.POD)), mdb.RENDER, arg[1:])
				}},
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(ctx.COMMAND, arg)
				}},
				cli.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if kit.Contains(arg[1], ";") {
					arg = kit.Split(arg[1], ";", ";", ";")
				}
				if m.Cmdy(m.Space(m.Option(cli.POD)), mdb.SEARCH, arg); arg[1] == "" {
					return
				}
				m.Cmd(mdb.INSERT, m.Prefix(P_SEARCH), "", mdb.HASH,
					kit.MDB_NAME, arg[1], kit.MDB_TYPE, arg[0], kit.MDB_TEXT, kit.Select("", arg, 2))
			}},
		}})
}
