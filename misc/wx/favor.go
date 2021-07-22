package wx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FAVOR: {Name: "favor", Help: "收藏", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_TEXT, kit.MDB_FIELD, "time,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			FAVOR: {Name: "favor text auto create", Help: "收藏", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(FAVOR), "", mdb.HASH, m.OptionSimple(kit.MDB_TEXT))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) == 0, m.Conf(FAVOR, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), "", mdb.HASH, kit.MDB_TEXT, arg)
				m.Table(func(index int, value map[string]string, head []string) {
					m.PushImages(cli.QRCODE, kit.MergeURL("https://open.weixin.qq.com/qr/code", aaa.USERNAME, value[kit.MDB_TEXT]))
				})
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
