package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	EXPIRE_SECONDS = "expire_seconds"
)
const SCAN = "scan"

func init() {
	Index.MergeCommands(ice.Commands{
		SCAN: {Name: "scan access hash auto", Help: "桌牌", Meta: kit.Merge(Meta(), kit.Dict(ice.CTX_TRANS, kit.Dict(html.VALUE, kit.Dict(
			"QR_STR_SCENE", "临时码", "QR_LIMIT_STR_SCENE", "永久码",
		)))), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type=QR_STR_SCENE,QR_LIMIT_STR_SCENE name*=1 text icons expire_seconds=3600 space index* args", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m.Spawn(), arg)
				res := SpidePost(m, QRCODE_CREATE, "action_name", m.Option(mdb.TYPE), "action_info.scene.scene_str", h, m.OptionSimple(EXPIRE_SECONDS))
				mdb.HashModify(m, mdb.HASH, h, mdb.LINK, kit.Value(res, web.URL), mdb.TIME, m.Time(kit.Format("%ss", kit.Select("60", m.Option(EXPIRE_SECONDS)))))
			}},
		}, mdb.ExportHashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,name,text,icons,space,index,args,type,link")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS).PushAction("").Option(ice.MSG_ACTION, "")
			} else if mdb.HashSelect(m, arg[1:]...); len(arg) > 1 {
				kit.If(m.Time() < m.Append(mdb.TIME), func() { m.PushQRCode(cli.QRCODE, m.Append(mdb.LINK)) })
			}
		}},
	})
}
