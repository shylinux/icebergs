package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	MONTH  = "month"
	INCOME = "income"
	TAX    = "tax"
)
const SALARY = "salary"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		SALARY: {Name: SALARY, Help: "工资", Value: kit.Data(
			mdb.SHORT, MONTH, mdb.FIELD, "month,company,amount,income,tax",
		)},
	}, Commands: ice.Commands{
		SALARY: {Name: "salary month auto create", Help: "工资", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create month company amount income tax 公积金 养老保险 医疗保险 生育保险 工伤保险 失业保险 企业公积金 企业养老保险 企业医疗保险 企业生育保险 企业工伤保险 企业失业保险", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
			amount, income, tax := 0, 0, 0
			m.Table(func(index int, value ice.Maps, head []string) {
				amount += kit.Int(value[AMOUNT])
				income += kit.Int(value[INCOME])
				tax += kit.Int(value[TAX])
			})
			m.StatusTime(AMOUNT, amount, INCOME, income, TAX, tax)
		}},
	}})
}
