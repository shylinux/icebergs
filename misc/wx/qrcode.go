package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	EXPIRE_SECONDS = "expire_seconds"
)
const QRCODE = "qrcode"

func init() {
	Index.MergeCommands(ice.Commands{
		QRCODE: {Name: "qrcode access hash auto", Help: "桌牌", Meta: Meta(), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type=QR_STR_SCENE,QR_LIMIT_STR_SCENE name*=1 text* icons expire_seconds=3600 space index* args", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m.Spawn(), arg)
				res := SpidePost(m, QRCODE_CREATE, "action_name", m.Option(mdb.TYPE), "action_info.scene.scene_str", h, m.OptionSimple(EXPIRE_SECONDS))
				mdb.HashModify(m, mdb.HASH, h, mdb.LINK, kit.Value(res, web.URL), mdb.TIME, m.Time(kit.Format("%ss", kit.Select("60", m.Option(EXPIRE_SECONDS)))))
			}},
		}, mdb.ExportHashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,name,text,icons,space,index,args,type,link")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS).PushAction("").Option(ice.MSG_ACTION, "")
			} else if mdb.HashSelect(m, arg[1:]...); len(arg) > 1 {
				kit.If(m.Time() < m.Append(mdb.TIME), func() { m.PushQRCode(QRCODE, m.Append(mdb.LINK)) })
			}
		}},
	})
}
