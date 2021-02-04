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
			FILES: {Name: FILES, Help: "文件", Value: kit.Data(kit.MDB_SHORT, "data")},
		},
		Commands: map[string]*ice.Command{
			FILES: {Name: "files hash auto upload", Help: "扫码", Action: map[string]*ice.Action{
				web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
					m.Cmdy(mdb.INSERT, FILES, "", mdb.HASH, "data", up[0], kit.MDB_NAME, up[1])
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, FILES, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,name,data", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, cmd, "", mdb.HASH, kit.MDB_HASH, arg)

				m.Table(func(index int, value map[string]string, head []string) {
					m.PushDownload(value[kit.MDB_NAME], "/share/cache/"+value["data"])
					if len(arg) > 0 {
						switch kit.Ext(value["name"]) {
						case "png":
							m.PushImages("image", "/share/cache/"+value["data"])
						}
					}
				})
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
