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

const (
	PRICE = "price"
)
const GOODS = "goods"

func init() {
	Index.MergeCommands(ice.Commands{
		GOODS: {Name: "goods list", Icon: "mall.png", Help: "商品", Meta: kit.Dict(
			ctx.TRANS, kit.Dict(html.INPUT, kit.Dict(mdb.TYPE, "单位", PRICE, "价格", AMOUNT, "总价")),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create zone* name* text price* count*=1 type*=件,个,份,斤 image*=4@img"},
			mdb.MODIFY: {Name: "modify zone* name* text price* count*=1 type*=件,个,份,斤 image*=4@img"},
			nfs.IMAGE:  {Name: "image image*=4@img", Help: "图片", Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, arg) }},
			ORDER:      {Name: "order count*=1", Help: "选购", Hand: func(m *ice.Message, arg ...string) {}},
		}, mdb.ExportHashAction(ctx.TOOLS, Prefix(ORDER), mdb.FIELD, "time,hash,zone,name,text,price,count,type,image")), Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 0 && m.IsMobileUA(), func() { m.OptionDefault(ice.MSG_FIELDS, "zone,name,price,count,type,text,hash,time,image") })
			mdb.HashSelect(m, arg...).PushAction(ORDER).Action("filter:text")
			web.PushPodCmd(m, "", arg...)
			ctx.DisplayLocal(m, "")
			ctx.Toolkit(m, "")
			m.Sort("zone,name")
		}},
	})
}
