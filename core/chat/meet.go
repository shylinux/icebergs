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
	Index.Register(&ice.Context{Name: MEET, Help: "遇见", Configs: map[string]*ice.Config{
		MISS: {Name: MISS, Help: "miss", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,照片,性别,年龄,身高,体重,籍贯,户口,学历,学校,职业,公司,年薪,资产,家境",
		)},
	}, Commands: map[string]*ice.Command{
		MISS: {Name: "miss name auto create", Help: "资料", Meta: kit.Dict("_trans", kit.Dict("name", "姓名")), Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create name 照片 性别 年龄 身高 体重 籍贯 户口 学历 学校 职业 公司 年薪 资产 家境", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			msg := m.Spawn()
			mdb.HashSelect(msg, arg...).Table(func(index int, value map[string]string, head []string) {
				value["照片"] = ice.Render(m, ice.RENDER_IMAGES, value["照片"], kit.Select("100", "400", msg.FieldsIsDetail()))
				m.Push(m.OptionFields(), value, kit.Split(msg.OptionFields()))
			})
		}},
	}}, nil, MISS)
}
