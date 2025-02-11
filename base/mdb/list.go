package mdb

import (
	"encoding/csv"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func _list_fields(m *ice.Message) []string {
	return kit.Split(kit.Select(LIST_FIELD, m.OptionFields()))
}
func _list_inputs(m *ice.Message, prefix, chain string, field, value string) {
	list := map[string]int{}
	defer func() {
		delete(list, "")
		kit.For(list, func(k string, i int) { m.Push(field, k).Push(COUNT, i) })
		m.SortIntR(COUNT)
	}()
	defer RLock(m, prefix)()
	Grows(m, prefix, chain, "", "", func(value ice.Map) {
		value = kit.GetMeta(value)
		list[kit.Format(value[field])] += kit.Int(kit.Select("1", value[COUNT]))
	})
}
func _list_insert(m *ice.Message, prefix, chain string, arg ...string) {
	m.Logs(INSERT, KEY, path.Join(prefix, chain), arg)
	defer Lock(m, prefix)()
	m.Echo("%d", Grow(m, prefix, chain, kit.Dict(arg, TARGET, m.Optionv(TARGET))))
	saveImportant(m, prefix, chain, kit.Simple(INSERT, prefix, chain, LIST, TIME, m.Time(), arg)...)
}
func _list_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Logs(MODIFY, KEY, path.Join(prefix, chain), field, value, arg)
	defer Lock(m, prefix)()
	Grows(m, prefix, chain, field, value, func(index int, val ice.Map) { _mdb_modify(m, val, field, arg...) })
	saveImportant(m, prefix, chain, kit.Simple(MODIFY, prefix, chain, LIST, field, value, arg)...)
}
func _list_select(m *ice.Message, prefix, chain, field, value string) {
	defer m.SortIntR(ID)
	fields := _list_fields(m)
	defer RLock(m, prefix)()
	Grows(m, prefix, chain, kit.Select(m.Option(CACHE_FIELD), field), kit.Select(m.Option(CACHE_VALUE), value), func(value ice.Map) {
		_mdb_select(m, m.OptionCB(""), "", value, fields, nil)
	})
}
func _list_export(m *ice.Message, prefix, chain, file string) {
	defer Lock(m, prefix)()
	p := kit.Keys(file, CSV)
	count := kit.Int(Conf(m, prefix, kit.Keys(chain, META, COUNT)))
	if count == 0 {
		if s, e := os.Stat(p); e == nil && !s.IsDir() {
			os.Remove(p)
		}
		return
	}
	f, p, e := miss.CreateFile(p)
	m.Assert(e)
	defer f.Close()
	defer m.Echo(p)
	m.Logs(EXPORT, KEY, path.Join(prefix, chain), FILE, p, COUNT, count)
	w := csv.NewWriter(f)
	defer w.Flush()
	head := kit.Split(ListField(m))
	Grows(m, prefix, chain, "", "", func(index int, value ice.Map) {
		if value = kit.GetMeta(value); index == 0 {
			kit.If(len(head) == 0 || head[0] == ice.FIELDS_DETAIL, func() { head = kit.SortedKey(value) })
			w.Write(head)
		}
		w.Write(kit.Simple(head, func(k string) string { return kit.Format(value[k]) }))
	})
	m.Conf(prefix, kit.Keys(chain, kit.Keym(COUNT)), 0)
	m.Conf(prefix, kit.Keys(chain, LIST), "")
}
func _list_import(m *ice.Message, prefix, chain, file string) {
	defer Lock(m, prefix)()
	f, e := miss.OpenFile(kit.Keys(file, CSV))
	if e != nil && !ice.Info.Important {
		return
	} else if m.WarnNotFound(e) {
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	head, _ := r.Read()
	count := 0
	for {
		line, e := r.Read()
		if e != nil {
			break
		}
		data := kit.Dict()
		for i, k := range head {
			if k == EXTRA {
				kit.Value(data, k, kit.UnMarshal(line[i]))
			} else {
				kit.Value(data, k, line[i])
			}
		}
		Grow(m, prefix, chain, data)
		count++
	}
	m.Logs(IMPORT, KEY, kit.Keys(prefix, chain), FILE, kit.Keys(file, CSV), COUNT, count)
	m.Echo("%d", count)
}

const (
	LIST_FIELD = "time,id,type,name,text"
)
const LIST = "list"

func ListAction(arg ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: AutoConfig(append(kit.List(FIELD, LIST_FIELD), arg...)...),
		INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(INPUTS, m.PrefixKey(), "", LIST, arg) }},
		INSERT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(INSERT, m.PrefixKey(), "", LIST, arg) }},
		DELETE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(DELETE, m.PrefixKey(), "", LIST, m.OptionSimple(ID), arg) }},
		MODIFY: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(MODIFY, m.PrefixKey(), "", LIST, m.OptionSimple(ID), arg) }},
		SELECT: {Name: "select id auto insert", Hand: func(m *ice.Message, arg ...string) { ListSelect(m, arg...) }},
		PRUNES: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(PRUNES, m.PrefixKey(), "", LIST, arg) }},
		EXPORT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(EXPORT, m.PrefixKey(), "", LIST, arg) }},
		IMPORT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(IMPORT, m.PrefixKey(), "", LIST, arg) }},
	}
}
func PageListAction(arg ...ice.Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		SELECT: {Name: "select id auto insert page", Hand: func(m *ice.Message, arg ...string) { PageListSelect(m, arg...) }},
		NEXT: {Hand: func(m *ice.Message, arg ...string) {
			NextPage(m, kit.Select(Config(m, COUNT), arg, 0), kit.Slice(arg, 1)...)
		}},
		PREV: {Hand: func(m *ice.Message, arg ...string) {
			PrevPageLimit(m, kit.Select(Config(m, COUNT), arg, 0), kit.Slice(arg, 1)...)
		}},
	}, ListAction(arg...))
}
func ListField(m *ice.Message) string { return kit.Select(LIST_FIELD, Config(m, FIELD)) }
func ListSelect(m *ice.Message, arg ...string) *ice.Message {
	m.Fields(len(kit.Slice(arg, 0, 1)), ListField(m))
	if m.Cmdy(SELECT, m.PrefixKey(), "", LIST, ID, arg); !m.FieldsIsDetail() {
		return m.StatusTimeCountTotal(Config(m, COUNT))
	}
	return m.StatusTime()
}
func PageListSelect(m *ice.Message, arg ...string) *ice.Message {
	OptionPage(m, kit.Slice(arg, 1)...)
	return ListSelect(m, arg...)
}
func NextPageLimit(m *ice.Message, total string, arg ...string) {
	if kit.Int(kit.Select("0", arg, 1)) > 0 {
		NextPage(m, total, arg...)
	} else {
		m.ProcessHold("已经是最后一页啦!")
	}
}
func NextPage(m *ice.Message, total string, arg ...string) {
	limit, offend := kit.Select("10", arg, 0), kit.Select("0", arg, 1)
	if offends := kit.Int(offend) - kit.Int(limit); total != "0" && (offends <= -kit.Int(total) || offends >= kit.Int(total)) {
		m.ProcessHold("已经是最后一页啦!")
	} else if offends == 0 {
		m.ProcessRewrite("offend", "")
	} else {
		m.ProcessRewrite("offend", offends)
	}
}
func PrevPage(m *ice.Message, total string, arg ...string) {
	limit, offend := kit.Select("10", arg, 0), kit.Select("0", arg, 1)
	if offends := kit.Int(offend) + kit.Int(limit); total != "0" && (offends <= -kit.Int(total) || offends >= kit.Int(total)) {
		m.ProcessHold("已经是最前一页啦!")
	} else if offends == 0 {
		m.ProcessRewrite("offend", "")
	} else {
		m.ProcessRewrite("offend", offends)
	}
}
func PrevPageLimit(m *ice.Message, total string, arg ...string) {
	if kit.Int(kit.Select("0", arg, 1)) < 0 {
		PrevPage(m, total, arg...)
	} else {
		m.ProcessHold("已经是最前一页啦!")
	}
}

func OptionPages(m *ice.Message, arg ...string) (page int, size int) {
	m.Option(CACHE_OFFEND, kit.Select(m.Option(CACHE_OFFEND), arg, 0))
	m.Option(CACHE_LIMIT, kit.Select(m.Option(CACHE_LIMIT), arg, 1))
	m.Option(CACHE_FILTER, kit.Select(m.Option(CACHE_FILTER), arg, 2))
	m.Option(OFFEND, kit.Select(m.Option(OFFEND), arg, 0))
	m.Option(LIMIT, kit.Select(m.Option(LIMIT), arg, 1))
	size = kit.Int(kit.Select("10", m.Option(LIMIT)))
	page = kit.Int(m.Option(OFFEND))/size + 1
	return
}
func OptionPage(m *ice.Message, arg ...string) int {
	page, _ := OptionPages(m, arg...)
	return page
}

const (
	CACHE_LIMIT  = "cache.limit"
	CACHE_BEGIN  = "cache.begin"
	CACHE_COUNT  = "cache.count"
	CACHE_OFFEND = "cache.offend"
	CACHE_FILTER = "cache.filter"
	CACHE_VALUE  = "cache.value"
	CACHE_FIELD  = "cache.field"
)

func Grows(m *ice.Message, prefix string, chain Any, match string, value string, cb Any) Map {
	cache, ok := m.Confv(prefix, chain).(ice.Map)
	if cache == nil || !ok {
		return nil
	} else if begin, limit := kit.Int(m.Option(CACHE_BEGIN)), kit.Int(m.Option(CACHE_LIMIT)); begin != 0 && limit > 0 {
		if count := kit.Int(m.Option(CACHE_COUNT, kit.Int(kit.Value(cache, kit.Keym(COUNT))))); count-begin < limit {
			m.Option(CACHE_OFFEND, 0)
			m.Option(CACHE_LIMIT, count-begin+1)
		} else {
			m.Option(CACHE_OFFEND, count-begin-limit+1)
		}
	}
	return miss.Grows(path.Join(prefix, kit.Keys(chain)), cache,
		kit.Int(kit.Select("0", strings.TrimPrefix(m.Option(CACHE_OFFEND), "-"))),
		kit.Int(kit.Select("10", m.Option(CACHE_LIMIT))), match, value, cb)
}
func Grow(m *ice.Message, prefix string, chain Any, data Any) int {
	cache, ok := m.Confv(prefix, chain).(ice.Map)
	if cache == nil || !ok {
		cache = kit.Data()
		m.Confv(prefix, chain, cache)
	}
	return miss.Grow(path.Join(prefix, kit.Keys(chain)), cache, data)
}
