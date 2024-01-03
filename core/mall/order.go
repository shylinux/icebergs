package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	ORDER_CONFIRM  = "confirm"
	ORDER_CANCEL   = "cancel"
	ORDER_PAYED    = "payed"
	ORDER_TRANSIT  = "transit"
	ORDER_RECEIVED = "received"
	ORDER_RETURNED = "returned"
	ORDER_REFUNDED = "refunded"

	PAY      = "pay"
	CANCEL   = "cancel"
	DELIVERY = "delivery"
	RECEIVE  = "receive"
	RETURN   = "return"
	REFUND   = "refund"
)
const ORDER = "order"

func init() {
	Index.MergeCommands(ice.Commands{
		ORDER: {Help: "订单", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
				m.Options(mdb.SUBKEY, kit.Keys(mdb.HASH, m.Option(mdb.HASH)))
				mdb.HashCreate(m, arg)
			}}, // mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {}},
			PAY:      {Help: "支付", Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, mdb.STATUS, ORDER_PAYED) }},
			CANCEL:   {Help: "取消", Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, mdb.STATUS, ORDER_CANCEL) }},
			DELIVERY: {Help: "发货", Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, mdb.STATUS, ORDER_TRANSIT) }},
			RECEIVE:  {Help: "收货", Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, mdb.STATUS, ORDER_RECEIVED) }},
			RETURN:   {Help: "退货", Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, mdb.STATUS, ORDER_RETURNED) }},
			REFUND:   {Help: "退钱", Hand: func(m *ice.Message, arg ...string) { mdb.HashModify(m, mdb.STATUS, ORDER_REFUNDED) }},
		}, mdb.ExportHashAction(mdb.FIELD, "time,hash,username,status,amount", mdb.FIELDS, "time,goods,price,count,units,name,text,image")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				stats := map[string]int{}
				mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
					switch value[mdb.STATUS] {
					case ORDER_CONFIRM:
						m.PushButton(PAY, CANCEL)
					case ORDER_CANCEL:
						m.PushButton(mdb.REMOVE)
					case ORDER_PAYED:
						m.PushButton(DELIVERY)
					case ORDER_TRANSIT:
						m.PushButton(RECEIVE)
					case ORDER_RECEIVED:
						m.PushButton(RETURN)
					case ORDER_RETURNED:
						m.PushButton(REFUND)
					default:
						m.PushButton("")
					}
					stats[value[mdb.STATUS]]++
				})
				m.StatusTimeCount(stats)
			} else {
				m.Options(mdb.SUBKEY, kit.Keys(mdb.HASH, arg[0])).OptionFields(mdb.Config(m, mdb.FIELDS))
				mdb.HashSelect(m, arg[1:]...)
				_status_amount(m)
			}
		}},
	})
}
