package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CART = "cart"

func init() {
	Index.MergeCommands(ice.Commands{
		CART: {Name: "cart list", Help: "购物车", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
				m.Options(mdb.SUBKEY, kit.Keys(mdb.HASH, mdb.HashCreate(m.Spawn(), aaa.USERNAME, m.Option(ice.MSG_USERNAME), mdb.SHORT, GOODS)))
				mdb.HashCreate(m.Spawn(), GOODS, m.Option(mdb.HASH), m.OptionSimple(mdb.COUNT))
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				m.Options(mdb.SUBKEY, kit.KeyHash(m.Option(ice.MSG_USERNAME)))
				mdb.HashRemove(m.Spawn(), m.OptionSimple(GOODS))
			}},
			ORDER: {Help: "下单", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd("")
				var amount float64
				msg.Table(func(value ice.Maps) {
					amount += kit.Float(value[PRICE]) * kit.Float(value[mdb.COUNT])
				})
				m.Options(mdb.HASH, m.Cmdx(ORDER, mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), mdb.STATUS, ORDER_CONFIRM, AMOUNT, amount))
				msg.Table(func(value ice.Maps) { m.Cmd(ORDER, mdb.INSERT, kit.Simple(value)) })
			}},
		}, mdb.ExportHashAction(mdb.SHORT, aaa.USERNAME, mdb.FIELD, "time,username", mdb.FIELDS, "time,goods,count")), Hand: func(m *ice.Message, arg ...string) {
			m.Options(mdb.SUBKEY, kit.KeyHash(m.Option(ice.MSG_USERNAME))).OptionFields(mdb.Config(m, mdb.FIELDS))
			mdb.HashSelect(m, arg...).Options(mdb.SUBKEY, "").Table(func(value ice.Maps) {
				m.Cmd(GOODS, value[GOODS], func(value ice.Maps) {
					m.Push("", value, kit.Split("name,text,price,units")).PushImages(nfs.IMAGE, web.SHARE_CACHE+value[nfs.IMAGE], "64")
				})
			}).Cut("image,name,text,price,count,units,goods,time").PushAction(mdb.DELETE).Action(ORDER)
			_status_amount(m)
		}},
	})
}
