package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const FILES = "files"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FILES: {Name: FILES, Help: "文件夹", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_DATA, kit.MDB_FIELD, "time,hash,type,name,size",
			)},
		},
		Commands: map[string]*ice.Command{
			FILES: {Name: "files hash auto upload", Help: "文件夹", Action: map[string]*ice.Action{
				web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
					m.Cmdy(mdb.INSERT, m.Prefix(FILES), "", mdb.HASH, kit.MDB_NAME, up[1], kit.MDB_TYPE, kit.Ext(up[1]), kit.MDB_DATA, up[0], kit.MDB_SIZE, up[2])
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(FILES), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(FILES, kit.META_FIELD))
				m.Cmd(mdb.SELECT, m.Prefix(FILES), "", mdb.HASH, kit.MDB_HASH, arg).Table(func(index int, value map[string]string, head []string) {
					link := kit.MergeURL("/share/cache/"+value[kit.MDB_DATA], "pod", m.Option(ice.MSG_USERPOD))
					m.Push("", value, kit.Split(m.Option(ice.MSG_FIELDS)))
					if m.PushDownload(kit.MDB_LINK, value[kit.MDB_NAME], link); len(arg) > 0 {
						switch {
						case kit.ExtIsImage(value[kit.MDB_NAME]):
							m.PushImages(kit.MDB_IMAGE, link)
						}
					}
				})
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
