package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	MEET = "meet"
)
const MISS = "miss"

func init() {
	Index.Register(&ice.Context{Name: MEET, Help: "遇见", Configs: ice.Configs{
		MISS: {Name: MISS, Help: "miss", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,照片,性别,年龄,身高,体重,籍贯,户口,学历,学校,职业,公司,年薪,资产,家境",
		)},
	}, Commands: ice.Commands{
		"monkey": {Name: "monkey total=888 count=9 run", Help: "猴子开箱子", Meta: kit.Dict("_trans", kit.Dict("name", "姓名")), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name 照片 性别 年龄 身高 体重 籍贯 户口 学历 学校 职业 公司 年薪 资产 家境", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			total := kit.Int(arg[0])
			count := kit.Int(arg[1])
			for i := 1; i <= total; i++ {
				list := []int{}
				for j := 1; j <= i/2; j++ {
					if i%j == 0 {
						list = append(list, j)
					}
				}
				list = append(list, i)
				if len(list) == count {
					m.Push("index", i)
					m.Push("count", len(list))
					m.Push("value", kit.Format(list))
				}
			}
			m.StatusTimeCount()
		}},
		MISS: {Name: "miss name auto create", Help: "资料", Meta: kit.Dict("_trans", kit.Dict("name", "姓名")), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name 照片 性别 年龄 身高 体重 籍贯 户口 学历 学校 职业 公司 年薪 资产 家境", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			msg := m.Spawn()
			mdb.HashSelect(msg, arg...).Tables(func(value ice.Maps) {
				value["照片"] = ice.Render(m, ice.RENDER_IMAGES, value["照片"], kit.Select("100", "400", msg.FieldsIsDetail()))
				m.Push(m.OptionFields(), value, kit.Split(msg.OptionFields()))
			})
		}},
	}}, nil, MISS)
}
