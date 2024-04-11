package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const HOUSE = "house"

func init() {
	const (
		AREA  = "area"
		ROOM  = "room"
		PRICE = "price"
		BUILD = "build"
	)
	Index.MergeCommands(ice.Commands{
		HOUSE: {Help: "房子", Icon: "Home.png", Meta: kit.Dict(ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
			mdb.ZONE, "区域", mdb.NAME, "小区", AREA, "面积", ROOM, "户型", PRICE, "总价", BUILD, "建成时间",
		))), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create zone* type* price* area* room* name* build text link image=9@img"},
		}, web.ExportCacheAction(nfs.IMAGE), mdb.ExportHashAction(mdb.FIELD, "time,hash,zone,type,price,area,room,name,build,text,link,image", mdb.SORT, "zone,name")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
			web.PushPodCmd(m, "", arg...)
		}},
	})
}
