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
			mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,type,name,text",
			mdb.LINK, "https://open.weixin.qq.com/qr/code",
		)},
	}, Commands: map[string]*ice.Command{
		FAVOR: {Name: "favor text:text auto create", Help: "收藏", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create type name text", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(index int, value map[string]string, head []string) {
				m.PushQRCode(mdb.SCAN, kit.MergeURL(m.Config(mdb.LINK), aaa.USERNAME, value[mdb.TEXT]))
			})
		}},
	}})
}
