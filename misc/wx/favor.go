package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FAVOR: {Name: "favor", Help: "收藏", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_TEXT, kit.MDB_FIELD, "time,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		FAVOR: {Name: "favor text auto create", Help: "收藏", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create type name text", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(index int, value map[string]string, head []string) {
				m.PushQRCode(kit.MDB_SCAN, kit.MergeURL("https://open.weixin.qq.com/qr/code", aaa.USERNAME, value[kit.MDB_TEXT]))
			})
		}},
	}})
}
