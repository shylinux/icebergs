package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const ADMIN = "admin"

func init() {
	Index.MergeCommands(ice.Commands{
		ADMIN: {Help: "管理", Actions: ice.MergeActions(ice.Actions{
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Append(mdb.HASH)) }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(GOODS, mdb.INPUTS, arg) }},
		}, ctx.ConfAction(ctx.TOOLS, Prefix(GOODS)), GOODS), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(GOODS, arg).PushAction(mdb.MODIFY, mdb.REMOVE)
			m.Action(mdb.CREATE)
			ctx.DisplayTable(m)
			_status_amount(m)
		}},
	})
}

func _status_amount(m *ice.Message) (amount float64) {
	m.Table(func(value ice.Maps) { amount += kit.Float(value[PRICE]) * kit.Float(value[mdb.COUNT]) })
	m.StatusTimeCount(AMOUNT, kit.Format("%0.2f", amount))
	return
}
