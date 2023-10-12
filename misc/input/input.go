package input

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	TEXT   = "text"
	CODE   = "code"
	WEIGHT = "weight"
)
const (
	WORD = "word"
	LINE = "line"
)

type input struct {
	ice.Zone
	insert string `name:"insert zone=person text code weight"`
	load   string `name:"load file=usr/wubi-dict/wubi86 zone=wubi86"`
	save   string `name:"save file=usr/wubi-dict/person zone=person"`
	list   string `name:"list method code auto load" help:"输入法"`
}

func (s input) Inputs(m *ice.Message, arg ...string) {
	switch s.Zone.Inputs(m, arg...); arg[0] {
	case nfs.FILE:
		m.Cmdy(nfs.DIR, "usr/wubi-dict/", nfs.PATH)
	case mdb.ZONE:
		m.Push(arg[0], path.Base(m.Option(nfs.FILE)))
	}
}
func (s input) Load(m *ice.Message, arg ...string) {
	if f, e := nfs.OpenFile(m, m.Option(nfs.FILE)); !m.Warn(e) {
		defer f.Close()
		lib := kit.Select(path.Base(m.Option(nfs.FILE)), m.Option(mdb.ZONE))
		m.Assert(nfs.RemoveAll(m, path.Join(mdb.Config(m, mdb.STORE), lib)))
		s.Zone.Remove(m, mdb.ZONE, lib)
		s.Zone.Create(m, kit.Simple(mdb.ZONE, lib, ctx.ConfigSimple(m.Message, mdb.LIMIT, mdb.LEAST, mdb.STORE, mdb.FSIZE))...)
		prefix := kit.Keys(mdb.HASH, m.Result())
		kit.For(f, func(s string) {
			if strings.HasPrefix(s, "# ") {
				return
			}
			line := kit.Split(s)
			if len(line) < 2 || (len(line) > 2 && line[2] == "0") {
				return
			}
			mdb.Grow(m.Message, m.PrefixKey(), prefix, kit.Dict(TEXT, line[0], CODE, line[1], WEIGHT, kit.Select("999999", line, 2)))
		})
		mdb.Conf(m, m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LIMIT)), 0)
		mdb.Conf(m, m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LEAST)), 0)
		m.Echo("%s: %d", lib, mdb.Grow(m.Message, m.PrefixKey(), prefix, kit.Dict(TEXT, "成功", CODE, "z", WEIGHT, "0")))
	}
}
func (s input) Save(m *ice.Message, arg ...string) (n int) {
	if f, p, e := nfs.CreateFile(m.Message, m.Option(nfs.FILE)); m.Assert(e) {
		defer f.Close()
		m.Option(mdb.CACHE_LIMIT, -2)
		for _, lib := range kit.Split(m.Option(mdb.ZONE)) {
			mdb.Richs(m.Message, m.PrefixKey(), "", lib, func(key string, value ice.Map) {
				mdb.Grows(m.Message, m.PrefixKey(), kit.Keys(mdb.HASH, key), "", "", func(index int, value ice.Map) {
					if value[CODE] != "z" {
						fmt.Fprintf(f, "%s %s %s\n", value[TEXT], value[CODE], value[WEIGHT])
						n++
					}
				})
			})
		}
		m.Logs(mdb.EXPORT, nfs.FILE, p, mdb.COUNT, n)
		m.Echo("%s: %d", p, n)
	}
	return
}
func (s input) List(m *ice.Message, arg ...string) {
	if len(arg) < 2 || arg[1] == "" {
		return
	}
	switch arg[0] {
	case LINE:
	case WORD:
		arg[1] = "^" + arg[1] + mdb.FS
	}
	res := m.Cmdx(cli.SYSTEM, "grep", "-rn", arg[1], mdb.Config(m, mdb.STORE))
	bio := csv.NewReader(bytes.NewBufferString(strings.Replace(res, nfs.DF, mdb.FS, -1)))
	for i := 0; i < kit.Int(10); i++ {
		if line, e := bio.Read(); e != nil {
			break
		} else if len(line) > 6 {
			m.Push(mdb.ID, line[3]).Push(CODE, line[2]).Push(TEXT, line[4]).Push(WEIGHT, line[6])
		}
	}
	m.SortIntR(WEIGHT)
}
