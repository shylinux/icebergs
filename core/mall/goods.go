package mall

import (
	"path"

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
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m.Spawn(kit.Dict(ice.MSG_FIELDS, nfs.IMAGE))).Table(func(value ice.Maps) {
					kit.For(kit.Split(value[nfs.IMAGE]), func(h string) {
						msg := m.Cmd(web.CACHE, h)
						m.Cmd(nfs.LINK, kit.Keys(path.Join(ice.USR_LOCAL_EXPORT, m.PrefixKey(), nfs.IMAGE, h), kit.Select("", kit.Split(msg.Append(mdb.TYPE), nfs.PS), -1)), msg.Append(nfs.FILE))
					})
				})
			}},
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				list := map[string]string{}
				m.Cmd(nfs.DIR, path.Join(ice.USR_LOCAL_EXPORT, m.PrefixKey(), nfs.IMAGE), func(value ice.Maps) {
					list[kit.TrimExt(value[nfs.PATH])] = m.Cmd(web.CACHE, web.CATCH, value[nfs.PATH]).Append(mdb.HASH)
				})
				mdb.HashSelectUpdate(m, "", func(value ice.Map) {
					value[nfs.IMAGE] = kit.Join(kit.Simple(kit.For(kit.Split(kit.Format(value[nfs.IMAGE])), func(p string) string { return kit.Select(p, list[p]) })))
				})
			}},
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
			var total float64
			m.Table(func(value ice.Maps) { total += kit.Float(value[PRICE]) * kit.Float(value[mdb.COUNT]) })
			m.StatusTimeCount(AMOUNT, kit.Format("%0.2f", total))
		}},
	})
}
