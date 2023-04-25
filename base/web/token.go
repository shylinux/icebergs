package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const TOKEN = "token"

func init() {
	Index.MergeCommands(ice.Commands{
		TOKEN: {Name: "token hash auto create prunes", Help: "令牌", Actions: ice.MergeActions(mdb.HashAction(mdb.EXPIRE, mdb.MONTH, mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.Cmdy("web.code.publish", ice.CONTEXTS, kit.Dict(TOKEN, arg[0]))
			}
		}},
	})
}
