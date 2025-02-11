package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	AMOUNT = "amount"
	PRICE  = "price"
	UNITS  = "units"
)
const GOODS = "goods"

func init() {
	Index.MergeCommands(ice.Commands{
		GOODS: {Help: "商品", Icon: "mall.png", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create zone* name* text price* count*=1 units*=件 image*=4@img"},
			// mdb.MODIFY: {Name: "modify zone* name* text price* count*=1 units*=件 image*=4@img"},
			ORDER: {Name: "order count*=1", Help: "选购", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(CART, mdb.INSERT, arg) }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case "units":
					m.Push(arg[0], kit.Split(nfs.TemplateText(m, arg[0])))
				}
			}},
			web.STATS_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn()); msg.Length() > 0 {
					amount := msg.TableAmount(func(value ice.Maps) float64 { return kit.Float(value[mdb.COUNT]) * kit.Float(value[PRICE]) })
					web.PushStats(m, kit.Keys(m.CommandKey(), AMOUNT), amount, "元", "商品总额")
					web.PushStats(m, kit.Keys(m.CommandKey(), mdb.COUNT), msg.Length(), "", "商品数量")
				}
			}},
		}, web.StatsAction(), web.ExportCacheAction(nfs.IMAGE), mdb.ExportHashAction(ctx.TOOLS, kit.Fields(Prefix(CART), Prefix(ORDER)), mdb.FIELD, "time,hash,zone,name,text,price,count,units,image")), Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 0 && m.IsMobileUA(), func() { m.OptionDefault(ice.MSG_FIELDS, "zone,name,price,count,units,text,hash,time,image") })
			mdb.HashSelect(m, arg...).PushAction(ORDER)
			web.PushPodCmd(m, "", arg...).Sort("zone,name")
			ctx.DisplayLocal(m, "")
			_status_amount(m)
		}},
	})
}
