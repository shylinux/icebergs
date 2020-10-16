package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const MALL = "mall"

var Index = &ice.Context{Name: MALL, Help: "贸易中心",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		"month": {Name: "month month value value 计算:button 记录:button", Help: "工资", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

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

func init() { web.Index.Register(Index, &web.Frame{}, ASSET) }
