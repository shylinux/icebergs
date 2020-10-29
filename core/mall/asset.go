package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _sub_key(m *ice.Message, account string) string {
	return kit.Keys(kit.MDB_HASH, kit.Hashs(account))
}
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
func _asset_list(m *ice.Message, account string, id string) {
	if account == "" {
		m.Option(mdb.FIELDS, "time,account,amount,count")
		defer m.PushAction(CHECK)
		defer m.SortIntR(AMOUNT)

	} else if id == "" {
		m.Option(mdb.FIELDS, "time,id,type,amount,name,text")

	} else {
		m.Option(mdb.FIELDS, mdb.DETAIL)
		defer m.PushAction(mdb.PLUGIN)
	}
	m.Cmdy(mdb.SELECT, ASSET, "", mdb.ZONE, account, id)
}

func _asset_check(m *ice.Message, account string) {
	amount := 0
	m.Option(mdb.FIELDS, "time,id,type,amount,name,text")
	m.Option(mdb.SELECT_CB, func(fields []string, value map[string]interface{}) {
		amount += kit.Int(kit.Value(value, AMOUNT))
	})
	m.Cmd(mdb.SELECT, ASSET, _sub_key(m, account), mdb.LIST)

	m.Cmdy(mdb.MODIFY, ASSET, "", mdb.HASH, ACCOUNT, account, AMOUNT, amount)
}
func _asset_create(m *ice.Message, account string) {
	m.Cmdy(mdb.INSERT, ASSET, "", mdb.HASH, ACCOUNT, account)
}
func _asset_insert(m *ice.Message, account string, arg ...string) {
	_asset_create(m, account)
	m.Cmdy(mdb.INSERT, ASSET, _sub_key(m, account), mdb.LIST, arg)

	m.Option(mdb.FIELDS, "time,account,amount,count")
	amount := kit.Int(m.Cmd(mdb.SELECT, ASSET, "", mdb.HASH, ACCOUNT, account).Append(AMOUNT))
	amount += kit.Int(_sub_value(m, AMOUNT, arg...))
	m.Cmdy(mdb.MODIFY, ASSET, "", mdb.HASH, ACCOUNT, account, AMOUNT, amount)
}
func _asset_modify(m *ice.Message, account, id string, arg ...string) {
	m.Cmdy(mdb.MODIFY, ASSET, _sub_key(m, account), mdb.LIST, kit.MDB_ID, id, arg)
}
func _asset_export(m *ice.Message, file string) {
	m.Option(mdb.FIELDS, "account,id,time,type,amount,name,text,extra")
	m.Cmdy(mdb.EXPORT, ASSET, "", mdb.ZONE, file)
}
func _asset_import(m *ice.Message, file string) {
	m.Option(mdb.FIELDS, "account")
	m.Cmdy(mdb.IMPORT, ASSET, "", mdb.ZONE, file)
}
func _asset_inputs(m *ice.Message, field, value string) {
	switch field {
	case "pod", "extra.pod":
		m.Cmdy(web.ROUTE)
	case "ctx", "extra.ctx":
		m.Cmdy(ctx.CONTEXT)
	case "cmd", "extra.cmd":
		m.Cmdy(ctx.CONTEXT, kit.Select(m.Option("ctx"), m.Option("extra.ctx")), ctx.COMMAND)
	case "arg", "extra.arg":

	case "from", "to", ACCOUNT:
		m.Cmdy(mdb.INPUTS, ASSET, "", mdb.HASH, ACCOUNT, value)
	default:
		m.Cmdy(mdb.INPUTS, ASSET, _sub_key(m, m.Option(ACCOUNT)), mdb.LIST, field, value)
	}
}

const (
	ACCOUNT = "account"
	AMOUNT  = "amount"

	SPEND = "spend"
	TRANC = "tranc"
	BONUS = "bonus"
	CHECK = "check"
)
const ASSET = "asset"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ASSET: {Name: ASSET, Help: "资产", Value: kit.Data(kit.MDB_SHORT, ACCOUNT)},
		},
		Commands: map[string]*ice.Command{
			ASSET: {Name: "asset account id auto spend tranc bonus", Help: "资产", Action: map[string]*ice.Action{
				SPEND: {Name: "spend account name amount time@date text", Help: "支出", Hand: func(m *ice.Message, arg ...string) {
					_sub_amount(m, arg)
					_asset_insert(m, arg[1], kit.Simple(kit.MDB_TYPE, "支出", arg[2:])...)
				}},
				TRANC: {Name: "tranc from to amount time@date text", Help: "转账", Hand: func(m *ice.Message, arg ...string) {
					_asset_insert(m, arg[3], kit.Simple(kit.MDB_TYPE, "转入", kit.MDB_NAME, arg[1], arg[4:])...)
					_sub_amount(m, arg)
					_asset_insert(m, arg[1], kit.Simple(kit.MDB_TYPE, "转出", kit.MDB_NAME, arg[3], arg[4:])...)
				}},
				BONUS: {Name: "bonus account name amount time@date text", Help: "收入", Hand: func(m *ice.Message, arg ...string) {
					_asset_insert(m, arg[1], kit.Simple(kit.MDB_TYPE, "收入", arg[2:])...)
				}},
				CHECK: {Name: "check account", Help: "核算", Hand: func(m *ice.Message, arg ...string) {
					_asset_check(m, m.Option(ACCOUNT))
				}},

				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_asset_modify(m, m.Option(ACCOUNT), m.Option(kit.MDB_ID), arg[0], arg[1])
				}},
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					_asset_export(m, m.Option(kit.MDB_FILE))
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_asset_import(m, m.Option(kit.MDB_FILE))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					_asset_inputs(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				}},

				mdb.PLUGIN: {Name: "plugin extra.pod extra.ctx extra.cmd extra.arg", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					_asset_modify(m, m.Option(ACCOUNT), m.Option(kit.MDB_ID), kit.Simple(kit.Dict(arg))...)
				}},
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == "run" {
						m.Cmdy(arg[1], arg[2:])
						return
					}
					if len(arg) > 0 {
						m.Cmdy(ctx.COMMAND, arg[0])
						return
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_asset_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
			}},
		},
	})
}
