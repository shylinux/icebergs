package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	PRICE = "price"
)
const GOODS = "goods"

func init() {
	Index.MergeCommands(ice.Commands{
		GOODS: {Name: "goods hash@keyboard place@province date@date name@key auto", Help: "商品", Actions: ice.MergeActions(ice.Actions{
			mdb.MODIFY: {Name: "modify zone type name text price count image=4@img"},
			mdb.CREATE: {Name: "modify zone type name text price count image=4@img"},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) { web.Upload(m) }},
			"copy": {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.CREATE, m.OptionSimple("zone,type,name,text,price,count,image"))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,zone,type,name,text,price,count,image")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || arg[0] == "" {
				m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
				ctx.DisplayLocal(m, "")
			} else {
				for _, p := range kit.Split(m.Append("image")) {
					m.EchoImages(web.MergeURL2(m, web.SHARE_CACHE+p))
				}
			}
			m.PushAction("copy", mdb.MODIFY, mdb.REMOVE)
		}},
	})
}
