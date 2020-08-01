package mall

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"encoding/csv"
	"os"
	"path"
	"strconv"
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
func _asset_create(m *ice.Message, account string) {
	if m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), account, nil) == nil {
		m.Conf(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_META, kit.MDB_SHORT), ACCOUNT)
		m.Rich(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.Data(ACCOUNT, account, AMOUNT, 0))
		m.Log_CREATE(ACCOUNT, account)
	}
}
func _asset_insert(m *ice.Message, account string, arg ...string) {
	m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), account, func(key string, value map[string]interface{}) {
		for i := 0; i < len(arg)-1; i += 2 {
			if arg[i] == "amount" {
				kit.Value(value, "meta.amount", kit.Int(kit.Value(value, "meta.amount"))+kit.Int(arg[i+1]))
			}
		}
		id := m.Grow(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), kit.Dict(
			kit.MDB_EXTRA, kit.Dict(),
			arg,
		))
		m.Log_INSERT(ACCOUNT, account, kit.MDB_ID, id, arg[0], arg[1])
		m.Echo("%d", id)
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
func _asset_input(m *ice.Message, key, value string) {
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

var _asset_inputs = kit.List(
	"_input", "text", "name", "account", "value", "@key",
	"_input", "select", "name", "type", "values", []interface{}{
		"支出", "收入",
	},
	"_input", "text", "name", "amount",
	"_input", "text", "name", "name", "value", "@key",
	"_input", "text", "name", "text", "value", "@key",
)

const (
	ASSET = "asset"
	BONUS = "bonus"
	SPEND = "spend"
)
const (
	AMOUNT  = "amount"
	ACCOUNT = "account"
	EXPORT  = "usr/export/web.mall.asset/"
)

var Index = &ice.Context{Name: "mall", Help: "贸易中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ASSET: {Name: ASSET, Help: "资产", Value: kit.Data(kit.MDB_SHORT, ACCOUNT)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			m.Cmd(mdb.SEARCH, mdb.CREATE, ASSET, ASSET, m.Prefix())
			m.Cmd(mdb.RENDER, mdb.CREATE, ASSET, ASSET, m.Prefix())
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(ASSET) }},
		ASSET: {Name: "asset account=auto id=auto auto 添加:button 导出:button 导入:button", Help: "资产", Meta: kit.Dict(
			mdb.INSERT, _asset_inputs,
		), Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert [key value]...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_asset_create(m, arg[1])
				_asset_insert(m, arg[1], arg[2:]...)
			}},
			mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				_asset_modify(m, m.Option(ACCOUNT), m.Option(kit.MDB_ID), arg[0], arg[1])
			}},
			mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				_asset_export(m, kit.Select(path.Join(EXPORT, m.Option(ice.MSG_DOMAIN), "list.csv"), arg, 0))
			}},
			mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				_asset_import(m, kit.Select(path.Join(EXPORT, m.Option(ice.MSG_DOMAIN), "list.csv"), arg, 0))
			}},

			"input": {Name: "input key value", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				_asset_input(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if _asset_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1)); len(arg) < 2 {
				m.Table(func(index int, value map[string]string, head []string) {
				})
			} else {
				m.Table(func(index int, value map[string]string, head []string) {
					if value["key"] == "status" {
						m.Push("key", "action")
						m.Push("value", _asset_action(m, value["value"]))
					}
				})
			}
			return
			if m.Option("_action") == "保存" {
				arg = []string{"action", "save"}
			}
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "modify":
					// 修改数据
					m.Richs(cmd, nil, m.Option("account"), func(key string, account map[string]interface{}) {
						m.Grows(cmd, kit.Keys("hash", key), "id", arg[5], func(index int, current map[string]interface{}) {
							m.Log(ice.LOG_MODIFY, "%s: %d %s: %s->%s", key, index, kit.Value(current, arg[2]), arg[2], arg[3])
							kit.Value(current, arg[2], arg[3])
						})
					})

				case "save":
					// 保存数据
					m.Option("cache.limit", -2)
					if f, p, e := kit.Create(kit.Select("usr/local/asset.csv", arg, 2)); m.Assert(e) {
						defer f.Close()

						w := csv.NewWriter(f)
						defer w.Flush()

						w.Write([]string{"时间", "收支类型", "账目分类", "备注", "金额", "账户"})
						m.Richs(cmd, nil, kit.Select("*", arg, 3), func(key string, account map[string]interface{}) {
							if kit.Format(kit.Value(account, "meta.account")) == "流水" {
								return
							}
							m.Grows(cmd, kit.Keys("hash", key), kit.Select("", arg, 4), kit.Select("", arg, 5), func(index int, current map[string]interface{}) {
								w.Write([]string{
									kit.Format(current["time"]),
									kit.Format(current["type"]),
									kit.Format(current["name"]),
									kit.Format(current["text"]),
									kit.Format(current["value"]),
									kit.Format(kit.Value(account, "meta.account")),
								})
							})
						})
						m.Log(ice.LOG_EXPORT, "%s", p)
						m.Cmdy(web.STORY, "catch", "csv", p)
					}

				case "load":
					// 加载数据
					m.CSV(m.Cmdx("nfs.cat", arg[2])).Table(func(index int, data map[string]string, head []string) {
						v, _ := strconv.ParseFloat(data["金额"], 64)
						for _, account := range []string{kit.Select(data["账户"], arg, 3), "流水"} {
							// amount := kit.Int(v * 100)
							item := kit.Dict(
								"type", data["收支类型"], "name", data["账目分类"], "text", data["备注"], "value", kit.Int(v),
								"time", data["时间"], "extra", kit.UnMarshal(data["extra"]),
							)

							if m.Richs(cmd, nil, account, nil) == nil {
								// 添加账户
								m.Log(ice.LOG_CREATE, "account: %s", account)
								m.Rich(cmd, nil, kit.Data("account", account, "amount", "0", "bonus", "0", "spend", "0"))
							}

							m.Richs(cmd, nil, account, func(key string, value map[string]interface{}) {
								// 账户流水
								m.Grow(cmd, kit.Keys("hash", key), item)

								// 账户结余
								amount := kit.Int(kit.Value(value, "meta.amount")) + kit.Int(v)
								m.Log(ice.LOG_INSERT, "%s: %v", key, amount)
								kit.Value(value, "meta.amount", amount)

								switch data["收支类型"] {
								case "收入":
									bonus := kit.Int(kit.Value(value, "meta.bonus")) + kit.Int(v)
									kit.Value(value, "meta.bonus", bonus)
								case "支出":
									spend := kit.Int(kit.Value(value, "meta.spend")) + kit.Int(v)
									kit.Value(value, "meta.spend", spend)
								}
							})
						}
					})
				}
				return
			}

			if len(arg) == 0 {
				// 账户列表
				m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"], []string{"account", "count", "amount", "bonus", "spend"})
				})
				m.Sort("amount", "int_r")
				return
			}

			if len(arg) > 5 && m.Richs(cmd, nil, arg[0], nil) == nil {
				// 添加账户
				m.Rich(cmd, nil, kit.Data("account", arg[0], "amount", "0", "bonus", "0", "spend", "0"))
				m.Log(ice.LOG_CREATE, "account: %s", arg[0])
			}

			field := []string{"time", "id", "value", "type", "name", "text"}
			m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
				if len(arg) == 1 {
					// 消费流水
					m.Grows(cmd, kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						m.Push("", value, field)
					})
					m.Sort("id", "int_r")
					return
				}
				if len(arg) == 2 {
					// 消费详情
					m.Grows(cmd, kit.Keys("hash", key), "id", arg[1], func(index int, value map[string]interface{}) {
						m.Push("detail", value)
					})
					return
				}
				if len(arg) < 6 {
					// 消费查询
					name, value := "type", arg[2]
					switch len(arg) {
					case 3:
						// 消费分类
						name, value = "type", arg[2]
					case 4:
						// 消费对象
						name, value = "name", arg[3]
					case 5:
						// 消费备注
						name, value = "text", arg[4]
					}
					m.Grows(cmd, kit.Keys("hash", key), name, value, func(index int, value map[string]interface{}) {
						m.Push("", value, field)
					})
					m.Sort("id", "int_r")
					return
				}

				// 词汇统计
				web.Count(m, cmd, "meta.word.type", arg[2])
				web.Count(m, cmd, "meta.word.name", arg[3])
				web.Count(m, cmd, "meta.word.text", arg[4])
				web.Count(m, cmd, "meta.word.value", strings.TrimPrefix(arg[5], "-"))

				// 数据结构
				amount := kit.Int(arg[5])
				extra := kit.Dict()
				data := kit.Dict(
					kit.MDB_TYPE, arg[2], kit.MDB_NAME, arg[3], kit.MDB_TEXT, arg[4],
					"value", amount, "extra", extra,
				)
				for i := 6; i < len(arg)-1; i += 2 {
					switch arg[i] {
					case kit.MDB_TIME:
						kit.Value(data, arg[i], arg[i+1])
					default:
						kit.Value(extra, arg[i], arg[i+1])
					}
				}
				// 添加流水
				n := m.Grow(cmd, kit.Keys(kit.MDB_HASH, key), data)

				// 账户结余
				total := kit.Int(kit.Value(value, "meta.amount")) + amount
				m.Log(ice.LOG_INSERT, "account: %s total: %v", arg[0], total)
				kit.Value(value, "meta.amount", total)
				m.Echo("%s: %d %d\n", arg[0], n, total)

				// 收支统计
				switch data["type"] {
				case "收入":
					bonus := kit.Int(kit.Value(value, "meta.bonus")) + amount
					kit.Value(value, "meta.bonus", bonus)
				case "支出":
					spend := kit.Int(kit.Value(value, "meta.spend")) + amount
					kit.Value(value, "meta.spend", spend)
				}
			})
		}},

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

				switch m.Option(ice.MSG_ACTION) {
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
