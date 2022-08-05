package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const FILES = "files"

func init() {
	Index.MergeCommands(ice.Commands{
		FILES: {Name: "files hash auto upload", Help: "文件夹", Actions: ice.MergeActions(ice.Actions{
			web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				up := web.Upload(m)
				mdb.HashCreate(m, mdb.TYPE, kit.Ext(up[1]), mdb.NAME, up[1], nfs.SIZE, up[2], mdb.DATA, up[0])
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.DATA, mdb.FIELD, "time,hash,type,name,size,data")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Tables(func(value ice.Maps) {
				link := web.SHARE_CACHE + value[mdb.DATA]
				if m.PushDownload(mdb.LINK, value[mdb.NAME], link); len(arg) > 0 && kit.ExtIsImage(value[mdb.NAME]) {
					m.PushImages("image", link)
				}
			})
		}},
	})
}
