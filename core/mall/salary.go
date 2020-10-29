package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const SALARY = "salary"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SALARY: {Name: SALARY, Help: "工资", Value: kit.Data(kit.MDB_SHORT, "month")},
		},
		Commands: map[string]*ice.Command{
			SALARY: {Name: "salary month auto create", Help: "工资", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create company month amount income tax 公积金 养老保险 医疗保险 生育保险 工伤保险 失业保险 企业公积金 企业养老保险 企业医疗保险 企业生育保险 企业工伤保险 企业失业保险", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, SALARY, "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, SALARY, "", mdb.HASH, "month", m.Option("month"), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, SALARY, "", mdb.HASH, "month", m.Option("month"))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, SALARY, "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,month,company,amount,income,tax", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, SALARY, "", mdb.HASH, "month", arg)
			}},
		},
	})
}
