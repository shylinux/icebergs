package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const (
	LONGITUDE = "longitude"
	LATITUDE  = "latitude"
	PROVINCE  = "province"
	CITY      = "city"
	DISTRICT  = "district"
	STREET    = "street"
)

const LOCATION = "location"

func init() {
	chat.Index.MergeCommands(ice.Commands{
		LOCATION: {Help: "地图", Icon: "Maps.png", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(EXPLORE, "周边", DIRECTION, "导航", DISTRICT, "地区", html.INPUT, kit.Dict(
				LONGITUDE, "经度", LATITUDE, "纬度", PROVINCE, "省份", CITY, "城市", DISTRICT, "区域", STREET, "街路",
			)),
		), Actions: ice.MergeActions(ice.Actions{
			chat.FAVOR_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(arg[0] == mdb.TYPE, func() { m.Push(arg[0], LOCATION) })
			}},
			chat.FAVOR_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == LOCATION, func() { m.PushButton(kit.Dict(LOCATION, "地图")) })
			}},
			chat.FAVOR_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == LOCATION, func() { ctx.ProcessField(m, m.PrefixKey(), []string{m.Option(mdb.TEXT)}, arg...) })
			}},
		}, chat.FavorAction(), mdb.ExportHashAction(web.VENDOR, AMAP, mdb.FIELD, "time,hash,type,name,text,longitude,latitude,extra")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] != ctx.ACTION {
				mdb.HashSelect(m, kit.Slice(arg, 0, 1)...)
				web.PushPodCmd(m, "", arg...)
			}
			m.Cmdy(mdb.Config(m, web.VENDOR), arg)
		}},
	})
}
