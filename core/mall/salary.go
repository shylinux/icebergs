package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const (
	MONTH  = "month"
	INCOME = "income"
	TAX    = "tax"
)
const SALARY = "salary"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SALARY: {Name: SALARY, Help: "工资", Value: kit.Data(
				kit.MDB_SHORT, MONTH, kit.MDB_FIELD, "time,month,company,amount,income,tax",
			)},
		},
		Commands: map[string]*ice.Command{
			SALARY: {Name: "salary month auto create", Help: "工资", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create company month amount income tax 公积金 养老保险 医疗保险 生育保险 工伤保险 失业保险 企业公积金 企业养老保险 企业医疗保险 企业生育保险 企业工伤保险 企业失业保险", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(SALARY), "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(SALARY), "", mdb.HASH, m.OptionSimple(MONTH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(SALARY), "", mdb.HASH, m.OptionSimple(MONTH))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, m.Prefix(SALARY), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(SALARY, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(SALARY), "", mdb.HASH, MONTH, arg)
				amount, income, tax := 0, 0, 0
				m.Table(func(index int, value map[string]string, head []string) {
					amount += kit.Int(value[AMOUNT])
					income += kit.Int(value[INCOME])
					tax += kit.Int(value[TAX])
				})
				m.StatusTime(AMOUNT, amount, INCOME, income, TAX, tax)
			}},
		},
	})
}
