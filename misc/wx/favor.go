package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		FAVOR: {Name: "favor", Help: "收藏", Value: kit.Data(
			mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,type,name,text",
			mdb.LINK, "https://open.weixin.qq.com/qr/code",
		)},
	}, Commands: ice.Commands{
		FAVOR: {Name: "favor text:text auto create", Help: "收藏", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type name text", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Tables(func(value ice.Maps) {
				m.PushQRCode(mdb.SCAN, kit.MergeURL(m.Config(mdb.LINK), aaa.USERNAME, value[mdb.TEXT]))
			})
		}},
	}})
}
