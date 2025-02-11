package web

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const PRODUCT = "product"

func init() {
	Index.MergeCommands(ice.Commands{
		PRODUCT: {Name: "product refresh", Help: "产品展示", Actions: mdb.HashAction(mdb.SHORT, "index", mdb.FIELD, "time,name,text,order,disable,index,args"), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).SortInt(mdb.ORDER)
		}},
	})
}

func AddPortalProduct(m *ice.Message, name, text string, order float64, arg ...string) {
	m.Cmd("web.product", mdb.CREATE, mdb.NAME, name, mdb.TEXT, strings.TrimSpace(text), mdb.ORDER, order, ctx.INDEX, m.PrefixKey(), ctx.ARGS, kit.Format(arg))
}
