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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FILES: {Name: FILES, Help: "文件夹", Value: kit.Data(
			mdb.SHORT, mdb.DATA, mdb.FIELD, "time,hash,type,name,size,data",
		)},
	}, Commands: map[string]*ice.Command{
		FILES: {Name: "files hash auto upload", Help: "文件夹", Action: ice.MergeAction(map[string]*ice.Action{
			web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
				if len(up) < 2 {
					msg := m.Cmdy(web.CACHE, web.UPLOAD)
					up = kit.Simple(msg.Append(mdb.HASH), msg.Append(mdb.NAME), msg.Append(nfs.SIZE))
				}
				m.Cmdy(mdb.INSERT, m.Prefix(FILES), "", mdb.HASH, mdb.TYPE, kit.Ext(up[1]), mdb.NAME, up[1], nfs.SIZE, up[2], mdb.DATA, up[0])
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.Table(func(index int, value map[string]string, head []string) {
				link := "/share/cache/" + value[mdb.DATA]
				if m.PushDownload(mdb.LINK, value[mdb.NAME], link); len(arg) > 0 && kit.ExtIsImage(value[mdb.NAME]) {
					m.PushImages("image", link)
				}
			})
		}},
	}})
}
