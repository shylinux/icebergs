package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"encoding/csv"
	"os"
	"strings"
)

func _asset_list(m *ice.Message, account string, id string, field ...interface{}) {
	fields := strings.Split(kit.Select("time,account,id,type,amount,name,text", m.Option("fields")), ",")
	m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Select(kit.MDB_FOREACH, account), func(key string, val map[string]interface{}) {
		if account == "" {
			m.Push(key, val["meta"], []string{kit.MDB_TIME, kit.MDB_COUNT, ACCOUNT, AMOUNT})
			return
		}
		if account = kit.Format(kit.Value(val, "meta.account")); id == "" {
			m.Grows(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				m.Push(account, value, fields, val["meta"])
			})
			m.Sort("id", "int_r")
			return
		}
		m.Grows(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Push("detail", value)
		})
	})
}
func _asset_modify(m *ice.Message, account, id, pro, set string) {
	m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Select(kit.MDB_FOREACH, account), func(key string, val map[string]interface{}) {
		m.Grows(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			switch pro {
			case ACCOUNT, kit.MDB_ID, kit.MDB_TIME:
				m.Info("not allow %v", key)
			default:
				m.Log_MODIFY(ACCOUNT, account, kit.MDB_ID, id, kit.MDB_KEY, pro, kit.MDB_VALUE, set)
				kit.Value(value, pro, set)
			}
		})
	})
}
func _asset_export(m *ice.Message, file string) {
	f, p, e := kit.Create(file)
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	m.Assert(w.Write([]string{
		ACCOUNT, kit.MDB_ID, kit.MDB_TIME,
		kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT,
		AMOUNT, kit.MDB_EXTRA,
	}))
	count := 0
	m.Option("cache.limit", -2)
	m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		m.Grows(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			m.Assert(w.Write(kit.Simple(
				kit.Format(kit.Value(val, "meta.account")),
				kit.Format(value[kit.MDB_ID]),
				kit.Format(value[kit.MDB_TIME]),
				kit.Format(value[kit.MDB_TYPE]),
				kit.Format(value[kit.MDB_NAME]),
				kit.Format(value[kit.MDB_TEXT]),
				kit.Format(value[AMOUNT]),
				kit.Format(value[kit.MDB_EXTRA]),
			)))
			count++
		})
	})
	m.Log_EXPORT("file", p, "count", count)
	m.Echo(p)
}
func _asset_import(m *ice.Message, file string) {
	f, e := os.Open(file)
	m.Assert(e)
	defer f.Close()

	r := csv.NewReader(f)
	heads, _ := r.Read()
	count := 0
	for {
		lines, e := r.Read()
		if e != nil {
			break
		}

		account := ""
		data := kit.Dict()
		for i, k := range heads {
			switch k {
			case ACCOUNT:
				account = lines[i]
			case kit.MDB_ID:
				continue
			case kit.MDB_EXTRA:
				kit.Value(data, k, kit.UnMarshal(lines[i]))
			default:
				kit.Value(data, k, lines[i])
			}
		}

		_asset_create(m, account)
		m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), account, func(key string, value map[string]interface{}) {
			kit.Value(value, "meta.amount", kit.Int(kit.Value(value, "meta.amount"))+kit.Int(data[AMOUNT]))

			id := m.Grow(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), data)
			m.Log_INSERT(ACCOUNT, account, kit.MDB_ID, id)
			count++
		})
	}
	m.Log_IMPORT("file", file, "count", count)
	m.Echo(file)
}
func _asset_inputs(m *ice.Message, key, value string) {
	switch key {
	case ACCOUNT:
		m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
			m.Push("account", kit.Value(val, "meta.account"))
			m.Push("count", kit.Select("0", kit.Format(kit.Value(val, "meta.count"))))
		})

	case "name", "text":
		list := map[string]int{}
		m.Option("cache.limit", 10)
		m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(k string, val map[string]interface{}) {
			m.Grows(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, k), "", "", func(index int, value map[string]interface{}) {
				list[kit.Format(value[key])]++
			})
		})
		for k, i := range list {
			m.Push("key", k)
			m.Push("count", i)
		}
	}
	m.Sort("count", "int_r")
}

func _asset_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		m.Grows(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			if value[kit.MDB_NAME] == name || strings.Contains(kit.Format(value[kit.MDB_TEXT]), name) {
				m.Push("pod", m.Option(ice.MSG_USERPOD))
				m.Push("ctx", m.Prefix())
				m.Push("cmd", ASSET)
				m.Push("time", value[kit.MDB_TIME])
				m.Push("size", 1)
				m.Push("type", ASSET)
				m.Push("name", value[kit.MDB_NAME])
				m.Push("text", kit.Format("%s:%d", kit.Value(val, "meta.zone"), kit.Int(value[kit.MDB_ID])))

			}
		})
	})
}
func _asset_render(m *ice.Message, kind, name, text string, arg ...string) {
	ls := strings.Split(text, ":")
	m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), ls[0], func(key string, val map[string]interface{}) {
		m.Grows(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "id", ls[1], func(index int, value map[string]interface{}) {
			m.Push("detail", value)
		})
	})
}
func _asset_action(m *ice.Message, status interface{}, action ...string) string {
	return strings.Join(action, "")
}

var _input_spend = kit.List(
	"_input", "text", "name", "account", "value", "@key",
	"_input", "select", "name", "type", "values", []interface{}{
		"支出", "转账", "收入",
	},
	"_input", "text", "name", "amount",
	"_input", "text", "name", "name", "value", "@key",
	"_input", "text", "name", "text", "value", "@key",
)
var _input_trans = kit.List(
	"_input", "text", "name", "account", "value", "@key",
	"_input", "select", "name", "type", "values", []interface{}{
		"转账", "支出", "收入",
	},
	"_input", "text", "name", "amount",
	"_input", "text", "name", "name", "value", "@key",
	"_input", "text", "name", "text", "value", "@key",
)
var _input_bonus = kit.List(
	"_input", "text", "name", "account", "value", "@key",
	"_input", "select", "name", "type", "values", []interface{}{
		"收入", "转账", "支出",
	},
	"_input", "text", "name", "amount",
	"_input", "text", "name", "name", "value", "@key",
	"_input", "text", "name", "text", "value", "@key",
)

const (
	BONUS = "bonus"
	SPEND = "spend"
)
const (
	AMOUNT  = "amount"
	ACCOUNT = "account"
	EXPORT  = "usr/export/web.mall.asset/"
)

var Index = &ice.Context{Name: "mall", Help: "贸易中心",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		"spend": {Name: "spend account=@key to=@key name=@key 记录:button text:textarea value=@key time=@date",
			Help: "支出", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if input(m, arg...) {
					// 输入补全
					return
				}
				if len(arg) < 2 {
					// 查看流水
					m.Cmdy("asset", arg)
					return
				}
				// 添加流水
				amount := kit.Int(arg[4])
				m.Cmdy("asset", arg[0], "", "转出", arg[1], arg[2], -amount, "time", arg[5:])
				m.Cmdy("asset", arg[1], "", "转入", arg[0], arg[2], amount, "time", arg[5:])
				m.Cmdy("asset", arg[1], "", "支出", arg[2], arg[3], -amount, "time", arg[5:])
				m.Cmdy("asset", "流水", "", "支出", arg[2], arg[3], -amount, "time", arg[5:])
			}},
		"trans": {Name: "trans account=@key to=@key name=@key 记录:button text:textarea value=@key time=@date",
			Help: "转账", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if input(m, arg...) {
					// 输入补全
					return
				}
				if len(arg) < 2 {
					// 查看流水
					m.Cmdy("asset", arg)
					return
				}
				// 添加流水
				amount := kit.Int(arg[4])
				m.Cmdy("asset", arg[0], "", "转出", arg[1], arg[2], -amount, "time", arg[5:])
				m.Cmdy("asset", arg[1], "", "转入", arg[0], arg[2], amount, "time", arg[5:])
				m.Cmd("asset", "流水", "", "转出", arg[2], arg[3], -amount, "time", arg[5:])
				m.Cmd("asset", "流水", "", "转入", arg[2], arg[3], amount, "time", arg[5:])
			}},
		"bonus": {Name: "bonus account=@key name=@key value=@key 记录:button text:textarea time=@date",
			Help: "收入", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if input(m, arg...) {
					// 输入补全
					return
				}
				if len(arg) < 2 {
					// 查看流水
					m.Cmdy("asset", arg)
					return
				}
				// 添加流水
				amount := kit.Int(arg[2])
				m.Cmdy("asset", arg[0], "", "收入", arg[1], arg[3], amount, "time", arg[4:])
				m.Cmdy("asset", "流水", "", "收入", arg[1], arg[3], amount, "time", arg[4:])
			}},
		"month": {Name: "month month value value 计算:button 记录:button",
			Help: "工资", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				// 输入补全
				if input(m, arg...) {
					return
				}

				data := map[string]int64{"个税方案": 6, "基本工资": 0, "个税": 0,
					"公积金比例": 1200, "养老保险比例": 800, "医疗保险比例": 200, "失业保险比例": 20, "工伤保险比例": 2, "生育保险比例": 0,
					"公积金金额": 0, "养老保险金额": 0, "医疗保险金额": 0, "失业保险金额": 0, "工伤保险金额": 0, "生育保险金额": 0,

					"企业公积金比例": 1200, "企业养老保险比例": 2000, "企业医疗保险比例": 1000, "企业失业保险比例": 100, "企业工伤保险比例": 30, "企业生育保险比例": 80,
					"企业公积金金额": 0, "企业养老保险金额": 0, "企业医疗保险金额": 0, "企业失业保险金额": 0, "企业工伤保险金额": 0, "企业生育保险金额": 0,
				}

				for i := 3; i < len(arg)-1; i += 2 {
					if _, ok := data[arg[i]]; ok {
						data[arg[i]] = kit.Int64(arg[i+1])
						arg[i] = ""
					}
				}

				data["养老保险比例"] = 725
				data["失业保险比例"] = 18

				salary := kit.Int64(arg[1])
				data["个税"] = kit.Int64(arg[2])
				base := data["基本工资"]
				if base == 0 {
					base = salary
				}

				// 五险一金
				amount := int64(0)
				for _, k := range []string{"公积金", "养老保险", "医疗保险", "失业保险", "工伤保险"} {
					m.Push("名目", k)
					value := -base * kit.Int64(data[k+"比例"]) / 10000
					m.Info("%v %v: %v %v", base, k, base*kit.Int64(data[k+"比例"]), value)
					if m.Push("比例", data[k+"比例"]); data[k+"金额"] == 0 {
						if k == "医疗保险" {
							value -= 300
						}
						data[k+"金额"] = value
					} else {
						value = data[k+"金额"]
					}
					amount += value
					m.Push("金额", data[k+"金额"])
				}

				// 企业五险一金
				company := int64(0)
				for _, k := range []string{"企业公积金", "企业养老保险", "企业医疗保险", "企业失业保险", "企业工伤保险", "企业生育保险"} {
					m.Push("名目", k)
					value := -base * kit.Int64(data[k+"比例"]) / 10000
					if m.Push("比例", data[k+"比例"]); data[k+"金额"] == 0 {
						data[k+"金额"] = value
					}
					company += -value
					m.Push("金额", data[k+"金额"])
				}
				m.Push("名目", "企业承担")
				m.Push("比例", "")
				m.Push("金额", company)

				// 其它津贴
				for i := 3; i < len(arg)-1; i += 2 {
					if arg[i] != "" && data[arg[i]] == 0 {
						m.Push("名目", arg[i])
						m.Push("比例", "")
						m.Push("金额", arg[i+1])
						amount += kit.Int64(arg[i+1])
					}
				}
				salary += amount

				m.Push("名目", "税前收入")
				m.Push("比例", "")
				m.Push("金额", salary)

				tax, amount := int64(0), salary
				if data["个税方案"] == 6 {
					// 2011年个税法案
					month := []int64{
						8350000, 4500,
						5850000, 3500,
						3850000, 3000,
						1250000, 2500,
						800000, 2000,
						500000, 1000,
						350000, 300,
					}

					for i := 0; i < len(month); i += 2 {
						if amount > month[i] {
							tax, amount = tax+(amount-month[i])*month[i+1]/10000, month[i]
						}
					}
					if data["个税"] != 0 {
						tax = data["个税"]
					}
					m.Push("名目", "个税")
					m.Push("比例", "")
					m.Push("金额", tax)

					m.Push("名目", "税后收入")
					m.Push("比例", "")
					m.Push("金额", salary-tax)
				} else {
					// 2019年个税法案
					// year := []int{
					// 	96000000, 4500,
					// 	66000000, 3500,
					// 	42000000, 3000,
					// 	30000000, 2500,
					// 	14400000, 2000,
					// 	3600000, 1000,
					// 	0, 300,
					// }
				}

				switch m.Option("_action") {
				case "计算":
				case "记录":
					// 收入
					m.Cmd("bonus", "工资", "企业承担", company, arg[0])
					m.Cmd("bonus", "工资", "基本工资", arg[1], arg[0])
					for i := 3; i < len(arg)-1; i += 2 {
						if arg[i] != "" && data[arg[i]] == 0 {
							m.Cmd("bonus", "工资", arg[i], arg[i+1], arg[0])
						}
					}

					// 转账
					m.Cmd("trans", "工资", "公积金", -data["企业公积金金额"], arg[0])
					for _, k := range []string{"企业养老保险", "企业医疗保险", "企业失业保险", "企业工伤保险", "企业生育保险"} {
						m.Cmd("trans", "工资", k, -data[k+"金额"], arg[0])
					}
					m.Cmd("trans", "工资", "公积金", -data["公积金金额"], arg[0])
					for _, k := range []string{"养老保险", "医疗保险", "失业保险"} {
						m.Cmd("trans", "工资", k, -data[k+"金额"], arg[0])
					}

					// 个税
					m.Cmd("trans", "工资", "个税", tax, arg[0])
				}
			}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
func input(m *ice.Message, arg ...string) bool {
	if len(arg) > 0 && arg[0] == "action" {
		switch arg[1] {
		case "input":
			switch arg[2] {
			case "account", "to":
				m.Richs("asset", nil, "*", func(key string, value map[string]interface{}) {
					m.Push(arg[2], kit.Value(value, "meta.account"))
					m.Push("count", kit.Value(value, "meta.count"))
				})
				m.Sort("count", "int_r")
				return true
			case "type", "name", "text", "value":
				m.Confm("asset", kit.Keys("meta.word", arg[2]), func(key string, value string) {
					m.Push(arg[2], key)
					m.Push("count", value)
				})
				m.Sort("count", "int_r")
				return true
			}
		}
	}
	return false
}
