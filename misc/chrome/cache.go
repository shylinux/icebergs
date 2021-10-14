package chrome

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CACHE = "cache"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CACHE: {Name: CACHE, Help: "爬虫缓存", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_LINK, kit.MDB_FIELD, "time,hash,step,size,total,type,name,text,link",
				kit.MDB_PATH, ice.USR_LOCAL_IMAGE,
			)},
		},
		Commands: map[string]*ice.Command{
			CACHE: {Name: "cache hash auto create prunes", Help: "爬虫缓存", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					if m.Cmdy(mdb.SELECT, m.Prefix(CACHE), "", mdb.HASH, m.OptionSimple(kit.MDB_LINK)); len(m.Appendv(kit.MDB_TOTAL)) > 0 {
						return // 已经下载
					}

					h := m.Cmdx(mdb.INSERT, m.Prefix(CACHE), "", mdb.HASH, m.OptionSimple("type,name,text,link"))
					value := kit.GetMeta(m.Confm(CACHE, kit.Keys(kit.MDB_HASH, h)))
					m.Option(kit.Keycb(web.DOWNLOAD), func(size, total int) {
						value[kit.MDB_TOTAL], value[kit.MDB_SIZE], value[kit.MDB_STEP] = total, size, kit.Format(size*100/total)
					})
					msg := m.Cmd("web.spide", ice.DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option(kit.MDB_LINK))

					p := path.Join(m.Conf(CACHE, kit.META_PATH), m.Option(kit.MDB_NAME))
					m.Cmdy(nfs.LINK, p, msg.Append(kit.MDB_FILE))
					m.Toast("下载成功")
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(CACHE), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, m.Prefix(CACHE), "", mdb.HASH, kit.MDB_STEP, "100")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(CACHE, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(CACHE), "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
