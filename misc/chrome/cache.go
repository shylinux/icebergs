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
	daemon
	short string `data:"link"`
	field string `data:"show,count,total,value,hash,type,name,link"`
	path  string `data:"usr/local/image/"`
	list  string `name:"list hash auto prunes" help:"缓存"`
}

func (s cache) Create(m *ice.Message, arg ...string) *ice.Message {
	if s.Hash.List(m, m.Option(mdb.LINK)); m.Length() > 0 {
		return m
	}
	m.Option(mdb.HASH, s.Hash.Create(m.Spawn(), m.OptionSimple("show,type,name,link")...))
	web.Download(m.Message, m.Option(mdb.LINK), func(count, total, value int) {
		s.Hash.Modify(m, kit.Simple(mdb.COUNT, count, mdb.TOTAL, total, mdb.VALUE, value)...)
	})
	name := kit.Keys(path.Base(m.Append(nfs.FILE)), path.Base(m.Append(mdb.TYPE)))
	m.Cmdy(nfs.LINK, path.Join(mdb.Config(m, nfs.PATH), name), m.Append(nfs.FILE))
	s.Hash.Modify(m, mdb.NAME, name)
	web.ToastSuccess(m.Message)
	return m
}
func (s cache) List(m *ice.Message, arg ...string) {
	s.Hash.List(m, arg...)
}
func init() { ice.CodeCtxCmd(cache{}) }
