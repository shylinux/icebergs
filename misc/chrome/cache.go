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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CACHE: {Name: CACHE, Help: "爬虫缓存", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_LINK, kit.MDB_FIELD, "time,hash,step,size,total,type,name,text,link",
			nfs.PATH, ice.USR_LOCAL_IMAGE,
		)},
	}, Commands: map[string]*ice.Command{
		CACHE: {Name: "cache hash auto prunes", Help: "爬虫缓存", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(kit.MDB_LINK)); m.Length() > 0 {
					return // 已经下载
				}

				h := m.Cmdx(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple("type,name,text,link"))
				value := kit.GetMeta(m.Confm(m.PrefixKey(), kit.Keys(kit.MDB_HASH, h)))
				msg := m.Cmd("web.spide", ice.DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option(kit.MDB_LINK), func(size, total int) {
					value[kit.MDB_TOTAL], value[kit.MDB_SIZE], value[kit.MDB_STEP] = total, size, kit.Format(size*100/total)
				})

				p := path.Join(m.Config(nfs.PATH), m.Option(kit.MDB_NAME))
				m.Cmdy(nfs.LINK, p, msg.Append(nfs.FILE))
				m.Toast("下载成功")
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.PRUNES, m.PrefixKey(), "", mdb.HASH, kit.MDB_STEP, "100")
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
