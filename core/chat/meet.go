package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const (
	MEET = "meet"
)
const MISS = "miss"

func init() {
	Index.Register(&ice.Context{Name: MEET, Help: "遇见",
		Configs: map[string]*ice.Config{
			MISS: {Name: MISS, Help: "miss", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,name,照片,性别,年龄,身高,体重,籍贯,户口,学历,学校,职业,公司,年薪,资产,家境",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

			MISS: {Name: "miss name auto create", Help: "资料", Meta: kit.Dict("_trans", kit.Dict("name", "姓名")), Action: ice.MergeAction(map[string]*ice.Action{
				mdb.CREATE: {Name: "create name 照片 性别 年龄 身高 体重 籍贯 户口 学历 学校 职业 公司 年薪 资产 家境", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(MISS), "", mdb.HASH, arg)
				}},
			}, mdb.HashAction(MISS)), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(MISS, kit.META_FIELD))
				m.Cmd(mdb.SELECT, m.Prefix(MISS), "", mdb.HASH, kit.MDB_NAME, arg).Table(func(index int, value map[string]string, head []string) {
					value["照片"] = kit.Format(`<img src="%s" height=%s>`, value["照片"], kit.Select("100", "400", m.Option(mdb.FIELDS) == mdb.DETAIL))
					m.Push("", value, kit.Split(m.Option(ice.MSG_FIELDS)))
				})
			}},
		},
	}, nil, MISS)
}
