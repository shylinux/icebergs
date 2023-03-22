package alpha

import (
	"path"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
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
	field string `data:"word,phonetic,translation,definition"`
	store string `data:"usr/local/export/"`
	fsize string `data:"300000"`
	limit string `data:"50000"`
	least string `data:"1000"`
	load  string `name:"load file=usr/word-dict/ecdict zone=ecdict"`
	list  string `name:"list method=word,line word auto load" help:"词典"`
}

func (s alpha) Load(m *ice.Message, arg ...string) {
	lib := kit.Select(path.Base(m.Option(nfs.FILE)), m.Option(mdb.ZONE))
	m.Assert(nfs.RemoveAll(m, path.Join(mdb.Config(m, mdb.STORE), lib)))
	s.Zone.Remove(m, mdb.ZONE, lib)
	s.Zone.Create(m, kit.Simple(mdb.ZONE, lib, ctx.ConfigSimple(m.Message, mdb.FIELD, mdb.LIMIT, mdb.LEAST, mdb.STORE, mdb.FSIZE))...)
	prefix := kit.Keys(mdb.HASH, m.Result())
	m.Cmd(mdb.IMPORT, m.PrefixKey(), prefix, mdb.LIST, m.Option(nfs.FILE))
	mdb.Conf(m, "", kit.Keys(prefix, kit.Keym(mdb.LIMIT)), 0)
	mdb.Conf(m, "", kit.Keys(prefix, kit.Keym(mdb.LEAST)), 0)
	m.Echo("%s: %d", lib, mdb.Grow(m, m.PrefixKey(), prefix, kit.Dict(WORD, ice.SP)))
}
func (s alpha) List(m *ice.Message, arg ...string) {
	if len(arg) < 2 || arg[1] == "" {
		m.Cmdy(cache{}, kit.Slice(arg, 1))
		return
	}
	switch arg[1] = strings.TrimSpace(arg[1]); arg[0] {
	case LINE:
	case WORD:
		if m.Cmdy(cache{}, kit.Slice(arg, 1)); m.Length() > 0 {
			return
		}
		defer func() { kit.If(m.Length() > 0, func() { m.Cmd(cache{}, mdb.CREATE, m.AppendSimple()) }) }()
		m.OptionFields(ice.FIELDS_DETAIL)
		arg[1] = "^" + arg[1] + ice.FS
	}
	wiki.CSV(m.Message.Spawn(), m.Cmdx(cli.SYSTEM, "grep", "-rih", arg[1], mdb.Config(m, mdb.STORE)), kit.Split(mdb.Config(m, mdb.FIELD))...).Tables(func(value ice.Maps) {
		kit.If(m.FieldsIsDetail(), func() { m.PushDetail(value, mdb.Config(m, mdb.FIELD)) }, func() { m.PushRecord(value, mdb.Config(m, mdb.FIELD)) })
	}).StatusTimeCount()
}

func init() { ice.WikiCtxCmd(alpha{}) }
