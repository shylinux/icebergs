package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const SEARCH = "search"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		web.P(SEARCH): {Name: "/search", Help: "搜索引擎", Actions: ctx.CmdAction(mdb.SHORT, mdb.NAME), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(m.Space(m.Option(ice.POD)), mdb.SEARCH, arg).StatusTimeCount()
		}},
	}})
}
