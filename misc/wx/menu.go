package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
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
		MENU: {Name: "menu access hash auto", Help: "菜单", Meta: kit.Merge(Meta(), kit.Dict(ice.CTX_TRANS, kit.Dict(html.VALUE, kit.Dict(
			"click", "点击", "view", "链接", "location_select", "定位",
			"scancode_waitmsg", "扫码上传", "scancode_push", "扫码",
			"pic_photo_or_album", "照片", "pic_sysphoto", "拍照", "pic_weixin", "相册",
		)))), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create scene*=main river*=1,2,3 storm*=1,2,3,4,5,6 type*=click,view,location_select,scancode_waitmsg,scancode_push,pic_photo_or_album,pic_sysphoto,pic_weixin name* text icons space index args"},
			mdb.UPDATE: {Name: "update scene*", Hand: func(m *ice.Message, arg ...string) {
				defer web.ToastProcess(m)()
				list := kit.Dict()
				m.Cmd("", m.Option(ACCESS), func(value ice.Maps) {
					if value[SCENE] == m.Option(SCENE) {
						key := kit.Keys("button", kit.Int(value[RIVER])-1)
						kit.If(value[STORM] != "1", func() { key = kit.Keys(key, "sub_button", kit.Int(value[STORM])-2) })
						kit.If(value[mdb.TYPE] == "view", func() {
							if value[mdb.TEXT] == "" {
								if value[web.SPACE] != "" {
									value[mdb.TEXT] = web.S(value[web.SPACE]) + web.C(value[ctx.INDEX])
								} else {
									value[mdb.TEXT] = web.C(value[ctx.INDEX])
								}
							}
							value[mdb.TEXT] = m.MergeLink(value[mdb.TEXT])
						})
						kit.Value(list, key, kit.Dict(mdb.TYPE, value[mdb.TYPE], mdb.NAME, value[mdb.NAME], mdb.KEY, value[mdb.HASH], web.URL, value[mdb.TEXT]))
					}
				})
				m.Echo(kit.Formats(SpidePost(m, MENU_CREATE, web.SPIDE_DATA, kit.Formats(list))))
				m.ProcessHold()
			}},
		}, mdb.ExportHashAction(mdb.SHORT, "scene,river,storm", mdb.FIELD, "time,hash,scene,river,storm,type,name,text,icons,space,index,args")), Hand: func(m *ice.Message, arg ...string) {
			m.Option("cache.limit", "-1")
			if len(arg) == 0 {
				m.Cmdy(ACCESS).PushAction("").Option(ice.MSG_ACTION, "")
			} else if mdb.HashSelect(m, arg[1:]...).Sort(mdb.Config(m, mdb.SHORT), ice.STR, ice.INT, ice.INT); len(arg) == 1 {
				m.Action(mdb.CREATE, mdb.UPDATE)
			}
		}},
	})
}
