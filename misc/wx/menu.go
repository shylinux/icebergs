package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	SCENE = "scene"
	RIVER = "river"
	STORM = "storm"
)
const MENU = "menu"

func init() {
	Index.MergeCommands(ice.Commands{
		MENU: {Name: "menu access hash auto", Help: "菜单", Meta: Meta(), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create scene*=main river*=1,2,3 storm*=1,2,3,4,5,6 type*=click,view,scancode_push,scancode_waitmsg,pic_sysphoto,pic_photo_or_album,pic_weixin,location_select name* text icons space index args"},
			mdb.UPDATE: {Name: "update scene*", Hand: func(m *ice.Message, arg ...string) {
				list := kit.Dict()
				m.Cmd("", m.Option(ACCESS), func(value ice.Maps) {
					if value[SCENE] == m.Option(SCENE) {
						key := kit.Keys("button", kit.Int(value[RIVER])-1)
						kit.If(value[STORM] != "1", func() { key = kit.Keys(key, "sub_button", kit.Int(value[STORM])-2) })
						kit.If(value[mdb.TYPE] == "view", func() { value[mdb.TEXT] = web.MergeLink(m, value[mdb.TEXT]) })
						kit.Value(list, key, kit.Dict(mdb.TYPE, value[mdb.TYPE], mdb.NAME, value[mdb.NAME], mdb.KEY, value[mdb.HASH], web.URL, value[mdb.TEXT]))
					}
				})
				m.Echo(kit.Formats(SpidePost(m, MENU_CREATE, web.SPIDE_DATA, kit.Formats(list))))
			}},
		}, mdb.ExportHashAction(mdb.SHORT, "scene,river,storm", mdb.FIELD, "time,hash,scene,river,storm,type,name,text,icons,space,index,args")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS).PushAction("").Option(ice.MSG_ACTION, "")
			} else {
				mdb.HashSelect(m, arg[1:]...).Sort(mdb.Config(m, mdb.SHORT), ice.STR, ice.INT, ice.INT).Action(mdb.CREATE, mdb.UPDATE)
			}
		}},
	})
}
