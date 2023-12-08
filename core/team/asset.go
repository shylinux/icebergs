package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _sub_value(m *ice.Message, key string, arg ...string) string {
	for i := 0; i < len(arg)-1; i += 2 {
		if arg[i] == key {
			return arg[i+1]
		}
	}
	return ""
}
func _sub_amount(m *ice.Message, arg []string) {
	for i := 0; i < len(arg)-1; i += 2 {
		if arg[i] == AMOUNT {
			if amount := kit.Float(arg[i+1]); amount > 0 {
				arg[i+1] = kit.Format(-amount)
			}
		}
	}
}
func _asset_check(m *ice.Message, account string) {
	var amount float64
	m.OptionCB(mdb.SELECT, func(key string, value ice.Map) { amount += kit.Float(kit.Value(value, AMOUNT)) })
	m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.ZONE, account, ice.OptionFields(mdb.ZoneField(m)))
	m.Cmd(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, ACCOUNT, account, AMOUNT, amount)
}
func _asset_insert(m *ice.Message, account string, arg ...string) {
	m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, ACCOUNT, account)
	m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.ZONE, account, arg)
	amount := kit.Float(m.Cmdv(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, ACCOUNT, account, AMOUNT))
	amount += kit.Float(_sub_value(m, AMOUNT, arg...))
	m.Cmd(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, ACCOUNT, account, AMOUNT, amount)
}

const (
	ACCOUNT = "account"
	AMOUNT  = "amount"
	COUNT   = "count"
	FROM    = "from"
	TO      = "to"

	ICOME  = "income"
	SPEND  = "spend"
	TRANS  = "trans"
	INVEST = "invest"
	CHECK  = "check"
)
const ASSET = "asset"

func init() {
	Index.MergeCommands(ice.Commands{
		ASSET: {Name: "asset account id auto", Help: "资产", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(
				html.INPUT, kit.Dict(
					ACCOUNT, "账户", AMOUNT, "金额", FROM, "转出", TO, "转入", mdb.NAME, "商家", mdb.TEXT, "备注",
				),
				html.VALUE, kit.Dict(
					INCOME, "收入", SPEND, "支出", TRANS, "转账", INVEST, "投资",
				),
			),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				if kit.IsIn(arg[0], FROM, TO) {
					back := arg[0]
					arg[0] = ACCOUNT
					defer func() { m.RenameAppend(ACCOUNT, back) }()
				}
				mdb.ZoneInputs(m, arg)
			}},
			mdb.CREATE: {Name: "create account* type*"},
			INCOME: {Name: "income account* amount* name* time text", Help: "收入", Hand: func(m *ice.Message, arg ...string) {
				_asset_insert(m, arg[1], kit.Simple(mdb.TYPE, m.ActionKey(), arg[2:])...)
			}},
			SPEND: {Name: "spend account* amount* name* time text", Help: "支出", Hand: func(m *ice.Message, arg ...string) {
				_sub_amount(m, arg)
				_asset_insert(m, arg[1], kit.Simple(mdb.TYPE, m.ActionKey(), arg[2:])...)
			}},
			TRANS: {Name: "trans from* to* amount* time text", Help: "转账", Hand: func(m *ice.Message, arg ...string) {
				_asset_insert(m, arg[3], kit.Simple(mdb.TYPE, TRANS, mdb.NAME, arg[1], arg[4:])...)
				_sub_amount(m, arg)
				_asset_insert(m, arg[1], kit.Simple(mdb.TYPE, TRANS, mdb.NAME, arg[3], arg[4:])...)
			}},
			INVEST: {Name: "invset account* amount* name* time text", Help: "投资", Hand: func(m *ice.Message, arg ...string) {
				_asset_insert(m, arg[1], kit.Simple(mdb.TYPE, m.ActionKey(), arg[2:])...)
			}},
			CHECK: {Help: "核算", Hand: func(m *ice.Message, arg ...string) {
				defer web.ToastProcess(m)()
				if m.Option(ACCOUNT) == "" {
					m.Cmd("", func(value ice.Maps) { _asset_check(m, value[ACCOUNT]) })
				} else {
					_asset_check(m, m.Option(ACCOUNT))
				}
			}},
			web.STATS_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn()); msg.Length() > 0 {
					amount := msg.TableAmount(func(value ice.Maps) float64 { return kit.Float(value[AMOUNT]) })
					web.PushStats(m, kit.Keys(m.CommandKey(), AMOUNT), amount, "元", "资产总额")
					web.PushStats(m, kit.Keys(m.CommandKey(), mdb.COUNT), msg.Length(), "", "资产数量")
				}
			}},
		}, web.StatsAction(), mdb.ExportZoneAction(mdb.SHORT, ACCOUNT, mdb.FIELD, "time,account,type,amount,count", mdb.FIELDS, "time,id,type,amount,name,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.ZoneSelect(m, arg...)
			web.PushPodCmd(m, "", arg...)
			if m.SortIntR(AMOUNT); len(arg) == 0 {
				m.Action(INCOME, SPEND, TRANS, INVEST, CHECK, mdb.CREATE)
			} else {
				m.Action(INCOME, SPEND, TRANS, INVEST, CHECK)
			}
			var amount, count float64
			m.Table(func(value ice.Maps) {
				amount += kit.Float(value[AMOUNT])
				kit.If(len(arg) == 0, func() { count += kit.Float(value[COUNT]) }, func() { count++ })
			})
			m.StatusTime(COUNT, count, AMOUNT, amount)
		}},
	})
}
