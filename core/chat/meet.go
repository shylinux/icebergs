package chat

import (
	"math/rand"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
)

const (
	MISS  = "miss"
	MEDIA = "media"
)

const MEET = "meet"

func init() {
	Index.Register(&ice.Context{Name: MEET, Help: "遇见",
		Configs: map[string]*ice.Config{
			MISS: {Name: MISS, Help: "miss", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

			MEDIA: {Name: "media name auto open snapshot", Help: "视频会议", Meta: kit.Dict(
				"display", "/plugin/local/chat/media.js",
			), Action: map[string]*ice.Action{
				tcp.OPEN: {Name: "open", Help: "打开摄相头", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(MISS), "", mdb.HASH)
				}},
				"snapshot": {Name: "snapshot", Help: "拍照", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(MISS), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(MISS), "", mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
			MISS: {Name: "miss name auto create", Help: "资料", Meta: kit.Dict(
				"trans", kit.Dict("name", "姓名"),
			), Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name 照片 性别 年龄 身高 体重 籍贯 户口 学历 学校 职业 公司 年薪 资产 家境", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(MISS), "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(MISS), "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(MISS), "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(MISS), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(MISS), "", mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,name,照片,性别,年龄,身高,体重,籍贯,户口,学历,学校,职业,公司,年薪,资产,家境", mdb.DETAIL, len(arg) > 0))
				m.Cmd(mdb.SELECT, m.Prefix(MISS), "", mdb.HASH, kit.MDB_NAME, arg).Table(func(index int, value map[string]string, head []string) {
					value["照片"] = kit.Format(`<img src="%s" height=%s>`, value["照片"], kit.Select("100", "400", m.Option(mdb.FIELDS) == mdb.DETAIL))
					m.Push("", value, kit.Split(m.Option(mdb.FIELDS)))
				})
			}},
			"random": {Name: "random max auto", Help: "随机", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				max := kit.Int(kit.Select("6", arg, 0))
				n := rand.Intn(max + 1)
				m.Echo(`<span style="font-size:48px">%s</span>`, cli.Color(m, cli.RED, kit.Format(n)))
				m.Status("time", m.Time(), "max", max)
			}},
		},
	}, nil)
}
