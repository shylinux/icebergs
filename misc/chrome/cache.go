package chrome

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

type cache struct {
	ice.Hash
	operate

	short string `data:"link"`
	field string `data:"show,count,total,value,hash,type,name,link"`
	path  string `data:"usr/local/image"`
	list  string `name:"list hash auto prunes" help:"缓存"`
}

func (c cache) Create(m *ice.Message, arg ...string) *ice.Message {
	if c.Hash.List(m, m.Option(mdb.LINK)); m.Length() > 0 {
		return m // 已经下载
	}

	m.Option(mdb.HASH, c.Hash.Create(m.Spawn(), m.OptionSimple("show,type,name,link")...).Result())
	msg := m.Cmd("web.spide", ice.DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option(mdb.LINK), func(count, total, value int) {
		c.Hash.Modify(m, kit.Simple(mdb.COUNT, count, mdb.TOTAL, total, mdb.VALUE, kit.Format(value))...)
	})
	m.Cmdy(nfs.LINK, path.Join(m.Config(nfs.PATH), m.Option(mdb.NAME)), msg.Append(nfs.FILE))
	m.ToastSuccess()
	return m
}
func (c cache) Prunes(m *ice.Message, arg ...string) {
	c.Hash.Prunes(m, mdb.VALUE, "100")
}
func (c cache) List(m *ice.Message, arg ...string) {
	c.Hash.List(m, arg...)
}

func init() { ice.CodeCtxCmd(cache{}) }
