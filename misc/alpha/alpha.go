package alpha

import (
	"os"
	"path"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	WORD = "word"
	LINE = "line"
)

type alpha struct {
	cache
	field string `data:"word,translation,definition"`
	store string `data:"usr/local/export/alpha"`
	fsize string `data:"300000"`
	limit string `data:"50000"`
	least string `data:"1000"`

	load string `name:"load file=usr/word-dict/ecdict name=ecdict" help:"词典"`
	list string `name:"list method=word,line word auto" help:"词典"`
}

func (a alpha) Load(m *ice.Message, arg ...string) {
	name := m.Option(mdb.NAME)
	// 清空数据
	meta := m.Confm(m.PrefixKey(), mdb.META)
	m.Assert(os.RemoveAll(path.Join(kit.Format(meta[mdb.STORE]), name)))
	m.Conf(m.PrefixKey(), name, "")

	// 缓存配置
	m.Conf(m.PrefixKey(), kit.Keys(name, mdb.META), kit.Dict(meta))
	m.Cmd(mdb.IMPORT, m.PrefixKey(), name, mdb.LIST, m.Option(nfs.FILE))

	// 保存词库
	m.Conf(m.PrefixKey(), kit.Keys(name, kit.Keym(mdb.LIMIT)), 0)
	m.Conf(m.PrefixKey(), kit.Keys(name, kit.Keym(mdb.LEAST)), 0)
	m.Echo("%s: %d", name, m.Grow(m.PrefixKey(), name, kit.Dict(WORD, ice.SP)))
}
func (a alpha) List(m *ice.Message, arg ...string) {
	if len(arg) < 2 || arg[1] == "" {
		m.Cmdy(a.cache, kit.Slice(arg, 1))
		return
	}

	// 搜索方法
	switch arg[1] = strings.TrimSpace(arg[1]); arg[0] {
	case LINE:
		m.OptionFields(m.Config(mdb.FIELD))
	case WORD:
		if m.Cmdy(a.cache, kit.Slice(arg, 1)); m.Length() > 0 {
			return
		}
		defer func() {
			if m.Length() > 0 {
				m.Cmd(a.cache.Create, m.AppendSimple())
			}
		}()

		m.OptionFields(mdb.DETAIL)
		arg[1] = "^" + arg[1] + ice.FS
	}

	// 搜索词汇
	msg := m.Cmd(cli.SYSTEM, "grep", "-rih", arg[1], m.Config(mdb.STORE))
	msg.CSV(msg.Result(), kit.Split(m.Config(mdb.FIELD))...).Table(func(index int, value map[string]string, head []string) {
		if m.FieldsIsDetail() {
			m.Push(mdb.DETAIL, value, kit.Split(m.Config(mdb.FIELD)))
			m.Push(mdb.TIME, m.Time())
			return
		}
		m.PushSearch(kit.SimpleKV("", arg[0], value[WORD], value["translation"]), value)
	})
	m.StatusTimeCount()
}

func init() { ice.Cmd("web.wiki.alpha.alpha", alpha{}) }
