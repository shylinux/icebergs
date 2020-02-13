package mall

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var Index = &ice.Context{Name: "mall", Help: "贸易中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"railway": {Name: "railway", Help: "12306", Value: kit.Data()},
		"asset": {Name: "asset", Help: "资产", Value: kit.Data(
			kit.MDB_SHORT, "account", "limit", "5000",
			"site", kit.Dict(
				"个税", "https://its.beijing.chinatax.gov.cn:8443/zmsqjl.html",
				"社保", "http://fuwu.rsj.beijing.gov.cn/csibiz/indinfo/index.jsp",
				"公积金", "https://grwsyw.gjj.beijing.gov.cn/ish/flow/menu/PPLGRZH0102?_r=0.6644871172745264",
			),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", kit.Keys(m.Cap(ice.CTX_FOLLOW), "json"))

			// m.Cmd(ice.CTX_CONFIG, "load", "mall.json")
			// if m.Richs(ice.WEB_SPIDE, nil, "12306", nil) == nil {
			// 	m.Cmd(ice.WEB_SPIDE, "add", "12306", "https://kyfw.12306.cn")
			// }
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", kit.Keys(m.Cap(ice.CTX_FOLLOW), "json"), kit.Keys(m.Cap(ice.CTX_FOLLOW), "asset"))

			// m.Cmd(ice.CTX_CONFIG, "save", "mall.json", "web.mall.railway")
		}},

		"spend": {Name: "spend", Help: "支出", List: kit.List(
			kit.MDB_INPUT, "text", "name", "account",
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "text", "name", "text",
			kit.MDB_INPUT, "text", "name", "value",
			kit.MDB_INPUT, "button", "name", "记录",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy("asset", arg)
				return
			}
			amount := kit.Int(arg[3])
			m.Cmdy("asset", arg[0], "", "支出", arg[1], arg[2], -amount, arg[4:])
			m.Cmdy("asset", "流水", "", "支出", arg[1], arg[2], -amount, arg[4:])
		}},
		"trans": {Name: "trans", Help: "转账", List: kit.List(
			kit.MDB_INPUT, "text", "name", "account",
			kit.MDB_INPUT, "text", "name", "to",
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "text", "name", "text",
			kit.MDB_INPUT, "text", "name", "value",
			kit.MDB_INPUT, "button", "name", "记录",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy("asset", arg)
				return
			}
			amount := kit.Int(arg[4])
			m.Cmdy("asset", arg[0], "", "转出", arg[2], arg[3], -amount, arg[5:])
			m.Cmd("asset", arg[1], "", "转入", arg[2], arg[3], amount, arg[5:])
			m.Cmd("asset", "流水", "", "转出", arg[2], arg[3], -amount, arg[5:])
			m.Cmd("asset", "流水", "", "转入", arg[2], arg[3], amount, arg[5:])
		}},
		"bonus": {Name: "bonus", Help: "收入", List: kit.List(
			kit.MDB_INPUT, "text", "name", "account",
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "text", "name", "text",
			kit.MDB_INPUT, "text", "name", "value",
			kit.MDB_INPUT, "button", "name", "记录",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy("asset", arg)
				return
			}
			m.Cmdy("asset", arg[0], "", "收入", arg[1:])
			m.Cmdy("asset", "流水", "", "收入", arg[1:])
		}},
		"month": {Name: "month", Help: "工资", List: kit.List(
			kit.MDB_INPUT, "text", "name", "month",
			kit.MDB_INPUT, "text", "name", "value",
			kit.MDB_INPUT, "button", "name", "计算",
			kit.MDB_INPUT, "button", "name", "记录",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			data := map[string]int{"个税方案": 6, "基本工资": 0, "个税": 0,
				"公积金比例": 1200, "养老保险比例": 800, "医疗保险比例": 200, "失业保险比例": 20, "工伤保险比例": 2, "生育保险比例": 0,
				"公积金金额": 0, "养老保险金额": 0, "医疗保险金额": 0, "失业保险金额": 0, "工伤保险金额": 0, "生育保险金额": 0,

				"企业公积金比例": 1200, "企业养老保险比例": 2000, "企业医疗保险比例": 1000, "企业失业保险比例": 100, "企业工伤保险比例": 30, "企业生育保险比例": 80,
				"企业公积金金额": 0, "企业养老保险金额": 0, "企业医疗保险金额": 0, "企业失业保险金额": 0, "企业工伤保险金额": 0, "企业生育保险金额": 0,
			}

			for i := 2; i < len(arg)-1; i += 2 {
				if _, ok := data[arg[i]]; ok {
					data[arg[i]] = kit.Int(arg[i+1])
					arg[i] = ""
				}
			}

			// data["养老保险比例"] = 725
			// data["失业保险比例"] = 20
			//
			salary := kit.Int(arg[1])
			base := data["基本工资"]
			if base == 0 {
				base = salary
			}

			// 五险一金
			amount := 0
			for _, k := range []string{"公积金", "养老保险", "医疗保险", "失业保险", "工伤保险"} {
				m.Push("名目", k)
				value := -base * kit.Int(data[k+"比例"]) / 10000
				if m.Push("比例", data[k+"比例"]); data[k+"金额"] == 0 {
					data[k+"金额"] = value
				} else {
					value = data[k+"金额"]
				}
				amount += value
				m.Push("金额", data[k+"金额"])
			}

			// 企业五险一金
			company := 0
			for _, k := range []string{"企业公积金", "企业养老保险", "企业医疗保险", "企业失业保险", "企业工伤保险", "企业生育保险"} {
				m.Push("名目", k)
				value := -base * kit.Int(data[k+"比例"]) / 10000
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
			for i := 2; i < len(arg)-1; i += 2 {
				if arg[i] != "" && data[arg[i]] == 0 {
					m.Push("名目", arg[i])
					m.Push("比例", "")
					m.Push("金额", arg[i+1])
					amount += kit.Int(arg[i+1])
				}
			}
			salary += amount

			m.Push("名目", "税前收入")
			m.Push("比例", "")
			m.Push("金额", salary)

			tax, amount := 0, salary
			if data["个税方案"] == 6 {
				// 2011年个税法案
				month := []int{
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
				m.Cmd("bonus", "工资", "企业承担", arg[0], company)
				m.Cmd("bonus", "工资", "基本工资", arg[0], arg[1])
				for i := 2; i < len(arg)-1; i += 2 {
					if arg[i] != "" {
						m.Cmd("bonus", "工资", arg[i], arg[0], arg[i+1])
					}
				}

				// 转账
				m.Cmd("trans", "工资", "公积金", "企业公积金", arg[0], -data["企业公积金金额"])
				for _, k := range []string{"企业养老保险", "企业医疗保险", "企业失业保险", "企业工伤保险", "企业生育保险"} {
					m.Cmd("trans", "工资", "社保", k, arg[0], -data[k+"金额"])
				}
				m.Cmd("trans", "工资", "公积金", "个人公积金", arg[0], -data["公积金金额"])
				for _, k := range []string{"养老保险", "医疗保险", "失业保险"} {
					m.Cmd("trans", "工资", "社保", k, arg[0], -data[k+"金额"])
				}

				// 个税
				m.Cmd("trans", "工资", "个税", "工资个税", arg[0], tax)
			}
		}},
		"asset": {Name: "asset account type name value", Help: []string{"资产",
			"action save file [account [key value]]",
			"action load file [account]",
		}, List: kit.List(
			kit.MDB_INPUT, "text", "name", "account", "action", "auto",
			kit.MDB_INPUT, "text", "name", "id", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cache.limit", "10000")
			if len(arg) == 0 {
				// 账户列表
				m.Richs("asset", nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"], []string{"account", "count", "amount", "bonus", "spend"})
				})
				m.Sort("amount", "int_r")
				return
			}

			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "modify":
					m.Richs("asset", nil, m.Option("account"), func(key string, account map[string]interface{}) {
						m.Grows("asset", kit.Keys("hash", key), "id", arg[5], func(index int, current map[string]interface{}) {
							kit.Value(current, arg[2], arg[3])
						})
					})

				case "save":
					m.Option("cache.limit", "10000")
					if f, p, e := kit.Create(arg[2]); m.Assert(e) {
						defer f.Close()

						w := csv.NewWriter(f)
						defer w.Flush()

						w.Write([]string{"时间", "收支类型", "账目分类", "备注", "金额", "账户"})
						m.Richs("asset", nil, kit.Select("*", arg, 3), func(key string, account map[string]interface{}) {
							if kit.Format(kit.Value(account, "meta.account")) == "流水" {
								return
							}
							m.Grows("asset", kit.Keys("hash", key), kit.Select("", arg, 4), kit.Select("", arg, 5), func(index int, current map[string]interface{}) {
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
					}

				case "load":
					m.Option("cache.limit", "10000")
					m.CSV(m.Cmdx("nfs.cat", arg[2])).Table(func(index int, data map[string]string, head []string) {
						v, _ := strconv.ParseFloat(data["金额"], 64)
						for _, account := range []string{kit.Select(data["账户"], arg, 3), "流水"} {
							// amount := kit.Int(v * 100)
							item := kit.Dict(
								"type", data["收支类型"], "name", data["账目分类"], "text", data["备注"], "value", kit.Int(v),
								"time", data["时间"], "extra", kit.UnMarshal(data["extra"]),
							)

							if m.Richs("asset", nil, account, nil) == nil {
								// 添加账户
								m.Log(ice.LOG_CREATE, "account: %s", account)
								m.Rich("asset", nil, kit.Data("account", account, "amount", "0", "bonus", "0", "spend", "0"))
							}

							m.Richs("asset", nil, account, func(key string, value map[string]interface{}) {
								// 账户流水
								m.Grow("asset", kit.Keys("hash", key), item)

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

			if m.Richs("asset", nil, arg[0], nil) == nil {
				// 添加账户
				m.Rich("asset", nil, kit.Data("account", arg[0], "amount", "0", "bonus", "0", "spend", "0"))
				m.Log(ice.LOG_CREATE, "account: %s", arg[0])
			}

			m.Richs("asset", nil, arg[0], func(key string, value map[string]interface{}) {
				field := []string{"time", "id", "value", "type", "name", "text"}
				if len(arg) == 1 {
					// 消费流水
					m.Grows("asset", kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						m.Push("", value, field)
					})
					m.Sort("time", "time_r")
					return
				}
				if len(arg) == 2 {
					// 消费详情
					m.Grows("asset", kit.Keys("hash", key), "id", arg[1], func(index int, value map[string]interface{}) {
						m.Push("detail", value)
					})
					m.Sort("time", "time_r")
					return
				}
				if len(arg) < 6 {
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
					m.Grows("asset", kit.Keys("hash", key), name, value, func(index int, value map[string]interface{}) {
						m.Push("", value, field)
					})
					m.Sort("time", "time_r")
					return
				}

				// 添加流水
				amount := kit.Int(arg[5])
				extra := map[string]interface{}{}
				data := kit.Dict(
					"type", arg[2], "name", arg[3], "text", arg[4], "value", amount, "extra", extra,
				)
				for i := 6; i < len(arg); i += 2 {
					if arg[i] == "time" {
						kit.Value(data, arg[i], arg[i+1])
					} else {
						kit.Value(extra, arg[i], arg[i+1])
					}
				}
				m.Grow("asset", kit.Keys("hash", key), data)

				// 账户结余
				amount = kit.Int(kit.Value(value, "meta.amount")) + amount
				m.Log(ice.LOG_INSERT, "%s: %v", key, amount)
				kit.Value(value, "meta.amount", amount)
				m.Echo("%d", amount)

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

		"passcode": &ice.Command{Name: "passcode", Help: "passcode", Meta: kit.Dict("active", "mall/input"), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			prefix := []string{ice.WEB_SPIDE, "12306"}
			if len(arg) == 0 {
				m.Cmd(prefix, "raw", "/passport/web/auth/uamtk-static", "form", "appid", "otn")
				m.Cmd(prefix, "raw", "GET", "/otn/HttpZF/GetJS")
				m.Cmd(prefix, "raw", "/otn/login/conf")

				m.Cmdy(prefix, "GET", fmt.Sprintf("/passport/captcha/captcha-image64?login_site=E&module=login&rand=sjrand"))
				return
			}

			switch arg[0] {
			case "check":
				m.Cmdy(prefix, "GET", fmt.Sprintf("/passport/captcha/captcha-check?login_site=E&rand=sjrand&answer=%s", arg[1]))
			case "login":
				m.Cmdy(prefix, "raw", "/passport/web/login", "form", "username", arg[1], "password", arg[2], "answer", arg[3], "appid", "otn")
			}
		}},
		"railway": &ice.Command{Name: "railway", Help: "12306", List: kit.List(
			kit.MDB_INPUT, "text", "name", "date", "value", "2020-01-22",
			kit.MDB_INPUT, "text", "name", "from", "value", "北京",
			kit.MDB_INPUT, "text", "name", "to", "value", "曲阜",
			kit.MDB_INPUT, "button", "name", "查询",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if !m.Confs("railway", "meta.site") {
				list := strings.Split(strings.TrimPrefix(m.Cmdx(ice.WEB_SPIDE, "12306", "raw", "GET", "/otn/resources/js/framework/station_name.js?station_version=1.9090"), "var statuion_names ='"), "|")
				for i := 0; i < len(list)-5; i += 5 {
					m.Conf("railway", kit.Keys("meta.site", list[i+1]), list[i+2])
				}
			}

			date := time.Now().Add(time.Hour * 24).Format("2006-01-02")
			if len(arg) > 0 {
				date, arg = arg[0], arg[1:]
			}
			from := "北京"
			if len(arg) > 0 {
				from, arg = arg[0], arg[1:]
			}
			from_code := m.Conf("railway", kit.Keys("meta.site", from))
			to := "曲阜"
			if len(arg) > 0 {
				to, arg = arg[0], arg[1:]
			}
			to_code := m.Conf("railway", kit.Keys("meta.site", to))

			m.Echo("%s->%s %s\n", from, to, date)

			if len(arg) > 0 {
				m.Cmdy(ice.WEB_SPIDE, "12306", "raw", "GET", fmt.Sprintf("/otn/czxx/queryByTrainNo?train_no=%s&from_station_telecode=%s&to_station_telecode=%s&depart_date=%s",
					arg[0], from_code, to_code, date))
				return
			}

			m.Cmd(ice.WEB_SPIDE, "12306", "GET", fmt.Sprintf("/otn/leftTicket/init?linktypeid=dc&fs=%s,%s&ts=%s,%s&date=%s&flag=N,N,Y",
				from, from_code, to, to_code, date))
			m.Cmd(ice.WEB_SPIDE, "12306", "GET", fmt.Sprintf("/otn/leftTicket/queryZ?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT",
				date, from_code, to_code)).Table(func(index int, value map[string]string, head []string) {
				kit.Fetch(kit.Value(kit.UnMarshal(value["data"]), "result"), func(index int, value string) {
					fields := strings.Split(value, "|")
					m.Push("车次", fields[3])
					m.Push("出发", fields[8])
					m.Push("到站", fields[9])
					m.Push("时长", fields[10])
					m.Push("二等座", fields[30])
					m.Push("一等座", fields[31])
				})
			})
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
