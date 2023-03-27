package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const (
	PRICE = "price"
)
const GOODS = "goods"

func init() {
	Index.MergeCommands(ice.Commands{
		GOODS: {Name: "goods hash@keyboard place@province date@date name@key count=_number@keyboard auto music", Help: "商品", Actions: ice.MergeActions(ice.Actions{
			mdb.MODIFY: {Name: "modify zone type name text price count image=4@img audio video"},
			mdb.CREATE: {Name: "modify zone type name text price count image=4@img audio video"},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) { web.Upload(m) }},
			"copy": {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.CREATE, m.OptionSimple("zone,type,name,text,price,count,image"))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,zone,type,name,text,price,count,image,audio,video")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || arg[0] == "" {
				m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
				m.PushAction("copy", mdb.MODIFY, mdb.REMOVE)
				ctx.DisplayLocal(m, "")
			} else {
				// for _, p := range kit.Split(m.Append("image")) {
				// 	m.EchoImages(web.MergeURL2(m, web.SHARE_CACHE+p))
				// }
				// m.PushAction("play", "stop", "copy", mdb.MODIFY, mdb.REMOVE)
			}
		}},
	})
}
