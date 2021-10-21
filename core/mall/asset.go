package mall

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
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
	m.Cmd(mdb.SELECT, m.Prefix(ASSET), "", mdb.ZONE, account, ice.OptionFields(m.Conf(ASSET, kit.META_FIELD)))

	m.Cmdy(mdb.MODIFY, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, account, AMOUNT, amount)
}
func _asset_create(m *ice.Message, account string) {
	m.Cmdy(mdb.INSERT, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, account)
}
func _asset_insert(m *ice.Message, account string, arg ...string) {
	_asset_create(m, account)
	m.Cmdy(mdb.INSERT, m.Prefix(ASSET), "", mdb.ZONE, account, arg)

	m.OptionFields("time,account,amount,count")
	amount := kit.Int(m.Cmd(mdb.SELECT, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, account).Append(AMOUNT))
	amount += kit.Int(_sub_value(m, AMOUNT, arg...))
	m.Cmdy(mdb.MODIFY, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, account, AMOUNT, amount)
}
func _asset_inputs(m *ice.Message, field, value string) {
	if cli.Inputs(m, field) {
		return
	}

	switch strings.TrimPrefix(field, "extra.") {
	case ACCOUNT, FROM, TO:
		m.Cmdy(mdb.INPUTS, m.Prefix(ASSET), "", mdb.HASH, ACCOUNT, value)
	default:
		m.Cmdy(mdb.INPUTS, m.Prefix(ASSET), "", mdb.ZONE, m.Option(ACCOUNT), field, value)
	}
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
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ASSET: {Name: ASSET, Help: "资产", Value: kit.Data(
				kit.MDB_SHORT, ACCOUNT, kit.MDB_FIELD, "time,id,type,amount,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			ASSET: {Name: "asset account id auto spend trans bonus check", Help: "资产", Meta: kit.Dict(
				"_trans", kit.Dict(ACCOUNT, "账户", AMOUNT, "金额", FROM, "转出", TO, "转入", "time", "时间", "name", "商家", "text", "备注"),
			), Action: map[string]*ice.Action{
				SPEND: {Name: "spend account name amount time@date text", Help: "支出", Hand: func(m *ice.Message, arg ...string) {
					_sub_amount(m, arg)
					_asset_insert(m, arg[1], kit.Simple(kit.MDB_TYPE, "支出", arg[2:])...)
				}},
				TRANS: {Name: "trans from to amount time@date text", Help: "转账", Hand: func(m *ice.Message, arg ...string) {
					_asset_insert(m, arg[3], kit.Simple(kit.MDB_TYPE, "转入", kit.MDB_NAME, arg[1], arg[4:])...)
					_sub_amount(m, arg)
					_asset_insert(m, arg[1], kit.Simple(kit.MDB_TYPE, "转出", kit.MDB_NAME, arg[3], arg[4:])...)
				}},
				BONUS: {Name: "bonus account name amount time@date text", Help: "收入", Hand: func(m *ice.Message, arg ...string) {
					_asset_insert(m, arg[1], kit.Simple(kit.MDB_TYPE, "收入", arg[2:])...)
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

				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(ASSET), "", mdb.ZONE, m.Option(ACCOUNT), m.Option(kit.MDB_ID), arg)
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(ACCOUNT, m.Conf(ASSET, kit.META_FIELD), kit.MDB_EXTRA)
					m.Cmdy(mdb.EXPORT, m.Prefix(ASSET), "", mdb.ZONE)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(ACCOUNT)
					m.Cmdy(mdb.IMPORT, m.Prefix(ASSET), "", mdb.ZONE)
					m.Cmdy(ASSET, CHECK)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					_asset_inputs(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				}},

				mdb.PLUGIN: {Name: "plugin extra.pod extra.ctx extra.cmd extra.arg", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(ASSET), "", mdb.ZONE, m.Option(ACCOUNT), m.Option(kit.MDB_ID), arg)
				}},
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(ctx.COMMAND, arg)
				}},
				ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				amount, count := 0, 0
				m.Fields(len(arg), "time,account,amount,count", m.Conf(ASSET, kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, m.Prefix(ASSET), "", mdb.ZONE, arg); len(arg) == 0 {
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
				m.StatusTime(AMOUNT, amount, COUNT, count)
			}},
		},
	})
}
