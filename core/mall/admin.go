package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const ADMIN = "admin"

func init() {
	Index.MergeCommands(ice.Commands{
		ADMIN: {Name: "admin hash auto", Help: "管理", Meta: kit.Dict(
			ctx.TRANS, kit.Dict(html.INPUT, kit.Dict(mdb.TYPE, "单位", PRICE, "价格", AMOUNT, "总价")),
		), Actions: ice.MergeActions(ice.Actions{
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Append(mdb.HASH)) }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(GOODS, mdb.INPUTS, arg) }},
		}, GOODS, ctx.ConfAction(ctx.TOOLS, Prefix(GOODS))), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(GOODS, arg).PushAction(mdb.MODIFY, mdb.REMOVE)
			kit.If(m.IsMobileUA(), func() { m.Action(mdb.CREATE) }, func() { m.Action(mdb.CREATE, "filter:text") })
			kit.If(len(arg) > 0, func() {
				kit.For(kit.Split(m.Append(nfs.IMAGE)), func(p string) {
					m.EchoImages(web.MergeURL2(m, web.SHARE_CACHE+p))
				})
			})
			ctx.DisplayTable(m)
			ctx.Toolkit(m, "")
			var total float64
			m.Table(func(value ice.Maps) { total += kit.Float(value[PRICE]) * kit.Float(value[mdb.COUNT]) })
			m.StatusTimeCount(AMOUNT, kit.Format("%0.2f", total))
		}},
	})
}
