package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
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
			if amount := kit.Int(arg[i+1]); amount > 0 {
				arg[i+1] = kit.Format(-amount)
			}
		}
	}
}

func _asset_check(m *ice.Message, account string) {
	amount := 0
	m.Option(kit.Keycb(mdb.SELECT), func(key string, value map[string]interface{}) {
		amount += kit.Int(kit.Value(value, AMOUNT))
	})
	m.Cmd(mdb.SELECT, m.Prefix(ASSET), "", mdb.ZONE, account, ice.OptionFields(m.Config(mdb.FIELD)))

	m.Cmdy(mdb.MODIFY, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, account, AMOUNT, amount)
}
func _asset_insert(m *ice.Message, account string, arg ...string) {
	m.Cmdy(mdb.INSERT, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, account)
	m.Cmdy(mdb.INSERT, m.Prefix(ASSET), "", mdb.ZONE, account, arg)

	amount := kit.Int(m.Cmd(mdb.SELECT, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, account, ice.Option{mdb.FIELDS, "time,account,amount,count"}).Append(AMOUNT))
	amount += kit.Int(_sub_value(m, AMOUNT, arg...))
	m.Cmdy(mdb.MODIFY, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, account, AMOUNT, amount)
}

const (
	ACCOUNT = "account"
	AMOUNT  = "amount"
	COUNT   = "count"

	FROM = "from"
	TO   = "to"

	SPEND = "spend"
	TRANS = "trans"
	BONUS = "bonus"
	CHECK = "check"
)
const ASSET = "asset"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ASSET: {Name: ASSET, Help: "资产", Value: kit.Data(
			mdb.SHORT, ACCOUNT, mdb.FIELD, "time,id,type,amount,name,text",
			kit.MDB_ALIAS, kit.Dict(FROM, ACCOUNT, TO, ACCOUNT),
		)},
	}, Commands: map[string]*ice.Command{
		ASSET: {Name: "asset account id auto spend trans bonus", Help: "资产", Meta: kit.Dict(
			"_trans", kit.Dict(ACCOUNT, "账户", AMOUNT, "金额", FROM, "转出", TO, "转入", "time", "时间", "name", "商家", "text", "备注"),
		), Action: ice.MergeAction(map[string]*ice.Action{
			SPEND: {Name: "spend account name amount time@date text", Help: "支出", Hand: func(m *ice.Message, arg ...string) {
				_sub_amount(m, arg)
				_asset_insert(m, arg[1], kit.Simple(mdb.TYPE, "支出", arg[2:])...)
			}},
			TRANS: {Name: "trans from to amount time@date text", Help: "转账", Hand: func(m *ice.Message, arg ...string) {
				_asset_insert(m, arg[3], kit.Simple(mdb.TYPE, "转入", mdb.NAME, arg[1], arg[4:])...)
				_sub_amount(m, arg)
				_asset_insert(m, arg[1], kit.Simple(mdb.TYPE, "转出", mdb.NAME, arg[3], arg[4:])...)
			}},
			BONUS: {Name: "bonus account name amount time@date text", Help: "收入", Hand: func(m *ice.Message, arg ...string) {
				_asset_insert(m, arg[1], kit.Simple(mdb.TYPE, "收入", arg[2:])...)
			}},
			CHECK: {Name: "check", Help: "核算", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ACCOUNT) == "" {
					m.Cmd(ASSET).Table(func(index int, value map[string]string, head []string) {
						_asset_check(m, value[ACCOUNT])
					})
					m.ProcessRefresh30ms()
				} else {
					_asset_check(m, m.Option(ACCOUNT))
				}
				m.Toast("核算成功")
			}},
		}, mdb.ZoneAction(), ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), "time,account,amount,count", m.Config(mdb.FIELD))
			amount, count := 0, 0
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(CHECK)
				m.SortIntR(AMOUNT)

				m.Table(func(index int, value map[string]string, head []string) {
					amount += kit.Int(value[AMOUNT])
					count += kit.Int(value[COUNT])
				})

			} else {
				m.PushAction(mdb.PLUGIN)

				m.Table(func(index int, value map[string]string, head []string) {
					amount += kit.Int(value[AMOUNT])
					count++
				})
			}
			m.StatusTime(COUNT, count, AMOUNT, amount)
		}},
	}})
}
