package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const FILES = "files"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FILES: {Name: FILES, Help: "文件夹", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_DATA, kit.MDB_FIELD, "time,hash,type,name,size,data",
			)},
		},
		Commands: map[string]*ice.Command{
			FILES: {Name: "files hash auto upload", Help: "文件夹", Action: ice.MergeAction(map[string]*ice.Action{
				web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
					m.Cmdy(mdb.INSERT, m.Prefix(FILES), "", mdb.HASH, kit.MDB_TYPE, kit.Ext(up[1]), kit.MDB_NAME, up[1], kit.MDB_SIZE, up[2], kit.MDB_DATA, up[0])
				}},
			}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(FILES, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(FILES), "", mdb.HASH, kit.MDB_HASH, arg)
				m.Table(func(index int, value map[string]string, head []string) {
					link := kit.MergeURL("/share/cache/"+value[kit.MDB_DATA], ice.POD, m.Option(ice.MSG_USERPOD))
					if m.PushDownload(kit.MDB_LINK, value[kit.MDB_NAME], link); len(arg) > 0 && kit.ExtIsImage(value[kit.MDB_NAME]) {
						m.PushImages(kit.MDB_IMAGE, link)
					}
				})
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
