package alpha

import (
	"path"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

const (
	WORD = "word"
	LINE = "line"
)

type alpha struct {
	ice.Zone
	field string `data:"word,translation,definition"`
	store string `data:"usr/local/export"`
	fsize string `data:"300000"`
	limit string `data:"50000"`
	least string `data:"1000"`

	load string `name:"load file=usr/word-dict/ecdict name=ecdict" help:"加载"`
	list string `name:"list method=word,line word auto" help:"词典"`
}

func (s alpha) Load(m *ice.Message, arg ...string) {
	// 清空数据
	lib := kit.Select(path.Base(m.Option(nfs.FILE)), m.Option(mdb.ZONE))
	m.Assert(nfs.RemoveAll(m.Message, path.Join(m.Config(mdb.STORE), lib)))
	s.Zone.Remove(m, mdb.ZONE, lib)
	s.Zone.Create(m, kit.Simple(mdb.ZONE, lib, m.ConfigSimple(mdb.FIELD, mdb.LIMIT, mdb.LEAST, mdb.STORE, mdb.FSIZE))...)
	prefix := kit.Keys(mdb.HASH, m.Result())

	// 加载配置
	m.Cmd(mdb.IMPORT, m.PrefixKey(), prefix, mdb.LIST, m.Option(nfs.FILE))

	// 保存词库
	m.Conf(m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LIMIT)), 0)
	m.Conf(m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LEAST)), 0)
	m.Echo("%s: %d", lib, mdb.Grow(m.Message, m.PrefixKey(), prefix, kit.Dict(WORD, ice.SP)))
}
func (s alpha) List(m *ice.Message, arg ...string) {
	if len(arg) < 2 || arg[1] == "" {
		m.Cmdy(cache{}, kit.Slice(arg, 1))
		return // 缓存列表
	}

	switch arg[1] = strings.TrimSpace(arg[1]); arg[0] {
	case LINE:
	case WORD:
		if m.Cmdy(cache{}, kit.Slice(arg, 1)); m.Length() > 0 {
			return // 查询缓存
		}
		defer func() {
			if m.Length() > 0 { // 写入缓存
				m.Cmd(cache{}, mdb.CREATE, m.AppendSimple())
			}
		}()

		// 精确匹配
		m.OptionFields(ice.FIELDS_DETAIL)
		arg[1] = "^" + arg[1] + ice.FS
	}

	// 搜索词汇
	wiki.CSV(m.Message, m.Cmdx(cli.SYSTEM, "grep", "-rih", arg[1], m.Config(mdb.STORE)), kit.Split(m.Config(mdb.FIELD))...).Tables(func(value ice.Maps) {
		if m.FieldsIsDetail() {
			m.PushDetail(value, m.Config(mdb.FIELD))
		} else {
			m.PushRecord(value, m.Config(mdb.FIELD))
		}
	}).StatusTimeCount()
}

func init() { ice.Cmd("web.wiki.alpha.alpha", alpha{}) }
