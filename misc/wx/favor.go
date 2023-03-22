package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor text:text auto create", Help: "收藏", Actions: mdb.HashAction(
			mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,type,name,text", mdb.LINK, "https://open.weixin.qq.com/qr/code",
		), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Tables(func(value ice.Maps) {
				m.PushQRCode(mdb.SCAN, kit.MergeURL(mdb.Config(m, mdb.LINK), aaa.USERNAME, value[mdb.TEXT]))
			})
		}},
	})
}
