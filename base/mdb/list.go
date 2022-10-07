package mdb

import (
	"encoding/csv"
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
	defer RLock(m, prefix, chain)()

	list := map[string]int{}
	Grows(m, prefix, chain, "", "", func(val ice.Map) {
		val = kit.GetMeta(val)
		list[kit.Format(val[field])] += kit.Int(kit.Select("1", val[COUNT]))
	})
	for k, i := range list {
		m.Push(field, k)
		m.Push(COUNT, i)
	}
	m.SortIntR(COUNT)
}
func _list_insert(m *ice.Message, prefix, chain string, arg ...string) {
	defer Lock(m, prefix, chain)()

	m.Logs(INSERT, KEY, path.Join(prefix, chain), arg[0], arg[1])
	if m.Optionv(TARGET) != nil && m.Option(TARGET) != "" {
		m.Echo("%d", Grow(m, prefix, chain, kit.Dict(arg, TARGET, m.Optionv(TARGET))))
	} else {
		m.Echo("%d", Grow(m, prefix, chain, kit.Dict(arg)))
	}
}
func _list_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	defer Lock(m, prefix, chain)()

	Grows(m, prefix, chain, field, value, func(index int, val ice.Map) {
		m.Logs(MODIFY, KEY, path.Join(prefix, chain), field, value, arg)
		_mdb_modify(m, val, field, arg...)
	})
}
func _list_select(m *ice.Message, prefix, chain, field, value string) {
	defer RLock(m, prefix, chain)()

	fields := _list_fields(m)
	Grows(m, prefix, chain, kit.Select(m.Option(CACHE_FIELD), field), kit.Select(m.Option(CACHE_VALUE), value), func(value ice.Map) {
		_mdb_select(m, m.OptionCB(""), "", value, fields, nil)
	})
}
func _list_export(m *ice.Message, prefix, chain, file string) {
	defer RLock(m, prefix, chain)()

	f, p, e := miss.CreateFile(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	count := 0
	head := kit.Split(m.Config(FIELD))
	Grows(m, prefix, chain, "", "", func(index int, val ice.Map) {
		if val = kit.GetMeta(val); index == 0 {
			if len(head) == 0 || head[0] == ice.FIELDS_DETAIL { // 默认表头
				for k := range val {
					head = append(head, k)
				}
				kit.Sort(head)
			}
			w.Write(head) // 输出表头
		}

		data := []string{}
		for _, k := range head {
			data = append(data, kit.Format(val[k]))
		}
		w.Write(data) // 输出数据
		count++
	})

	m.Logs(EXPORT, KEY, path.Join(prefix, chain), FILE, p, COUNT, count)
	m.Conf(prefix, kit.Keys(chain, kit.Keym(COUNT)), 0)
	m.Conf(prefix, kit.Keys(chain, LIST), "")
	m.Echo(p)
}
func _list_import(m *ice.Message, prefix, chain, file string) {
	defer RLock(m, prefix, chain)()

	f, e := miss.OpenFile(kit.Keys(file, CSV))
	m.Assert(e)
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

	m.Logs(IMPORT, KEY, kit.Keys(prefix, chain), COUNT, count)
	m.Echo("%d", count)
}

const (
	LIST_FIELD = "time,id,type,name,text"
)
const LIST = "list"

func ListAction(args ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: AutoConfig(args...),
		INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INPUTS, m.PrefixKey(), "", LIST, ListArgs(m, arg))
		}},
		INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSERT, m.PrefixKey(), "", LIST, ListArgs(m, arg))
		}},
		DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(DELETE, m.PrefixKey(), "", LIST, m.OptionSimple(ID), arg)
		}},
		MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(MODIFY, m.PrefixKey(), "", LIST, m.OptionSimple(ID), ListArgs(m, arg))
		}},
		SELECT: {Name: "select", Help: "列表", Hand: func(m *ice.Message, arg ...string) {
			ListSelect(m, arg...)
		}},
		PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(PRUNES, m.PrefixKey(), "", LIST, arg)
		}},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(EXPORT, m.PrefixKey(), "", LIST, arg)
		}},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(IMPORT, m.PrefixKey(), "", LIST, arg)
		}},
		PREV: {Name: "prev", Help: "上一页", Hand: func(m *ice.Message, arg ...string) {
			PrevPage(m, m.Config(COUNT), kit.Slice(arg, 1)...)
		}},
		NEXT: {Name: "next", Help: "下一页", Hand: func(m *ice.Message, arg ...string) {
			NextPageLimit(m, m.Config(COUNT), kit.Slice(arg, 1)...)
		}},
	}
}
func ListField(m *ice.Message) string { return kit.Select(LIST_FIELD, m.Config(FIELD)) }
func ListArgs(m *ice.Message, arg ...ice.Any) []string {
	return _mdb_args(m, ListField(m), arg...)
}
func ListSelect(m *ice.Message, arg ...string) *ice.Message {
	OptionPage(m, kit.Slice(arg, 1)...)
	m.Fields(len(kit.Slice(arg, 0, 1)), ListField(m))
	if m.Cmdy(SELECT, m.PrefixKey(), "", LIST, ID, arg); !m.FieldsIsDetail() {
		m.StatusTimeCountTotal(m.Config(COUNT))
	} else {
		m.StatusTime()
	}
	return m
}
func PrevPageLimit(m *ice.Message, total string, arg ...string) {
	if kit.Int(kit.Select("0", arg, 1)) > 0 {
		PrevPage(m, total, arg...)
	} else {
		m.ProcessHold("已经最前一页啦!")
	}
}
func PrevPage(m *ice.Message, total string, arg ...string) {
	limit, offend := kit.Select("10", arg, 0), kit.Select("0", arg, 1)
	offends := kit.Int(offend) - kit.Int(limit)
	if total != "0" && (offends <= -kit.Int(total) || offends >= kit.Int(total)) {
		m.ProcessHold("已经是最前一页啦!")
		return
	}
	if offends == 0 {
		m.ProcessRewrite("offend", "")
	} else {
		m.ProcessRewrite("offend", offends)
	}

}
func NextPage(m *ice.Message, total string, arg ...string) {
	limit, offend := kit.Select("10", arg, 0), kit.Select("0", arg, 1)
	offends := kit.Int(offend) + kit.Int(limit)
	if total != "0" && (offends <= -kit.Int(total) || offends >= kit.Int(total)) {
		m.ProcessHold("已经是最后一页啦!")
		return
	}
	if offends == 0 {
		m.ProcessRewrite("offend", "")
	} else {
		m.ProcessRewrite("offend", offends)
	}
}
func NextPageLimit(m *ice.Message, total string, arg ...string) {
	if kit.Int(kit.Select("0", arg, 1)) < 0 {
		NextPage(m, total, arg...)
	} else {
		m.ProcessHold("已经是最后一页啦!")
	}
}

func OptionPages(m *ice.Message, arg ...string) (page int, size int) {
	m.Option(CACHE_LIMIT, kit.Select(m.Option(CACHE_LIMIT), arg, 0))
	m.Option(CACHE_OFFEND, kit.Select(m.Option(CACHE_OFFEND), arg, 1))
	m.Option(CACHE_FILTER, kit.Select(m.Option(CACHE_FILTER), arg, 2))
	m.Option(LIMIT, kit.Select(m.Option(LIMIT), arg, 0))
	m.Option(OFFEND, kit.Select(m.Option(OFFEND), arg, 1))
	size = kit.Int(kit.Select("10", m.Option(LIMIT)))
	page = kit.Int(m.Option(OFFEND))/size + 1
	return
}
func OptionPage(m *ice.Message, arg ...string) int {
	page, _ := OptionPages(m, arg...)
	return page
}

const ( // CACHE
	CACHE_LIMIT  = "cache.limit"
	CACHE_BEGIN  = "cache.begin"
	CACHE_COUNT  = "cache.count"
	CACHE_OFFEND = "cache.offend"
	CACHE_FILTER = "cache.filter"
	CACHE_VALUE  = "cache.value"
	CACHE_FIELD  = "cache.field"
)

func Grow(m *ice.Message, prefix string, chain Any, data Any) int {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		cache = kit.Data()
		m.Confv(prefix, chain, cache)
	}
	return miss.Grow(path.Join(prefix, kit.Keys(chain)), cache, data)
}
func Grows(m *ice.Message, prefix string, chain Any, match string, value string, cb Any) Map {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		return nil
	}

	limit := kit.Int(m.Option(CACHE_LIMIT))
	if begin := kit.Int(m.Option(CACHE_BEGIN)); begin != 0 && limit > 0 {
		count := kit.Int(m.Option(CACHE_COUNT, kit.Int(kit.Value(cache, kit.Keym("count")))))
		if begin > 0 {
			m.Option(CACHE_OFFEND, count-begin-limit)
		} else {
			m.Option(CACHE_OFFEND, -begin-limit)
		}
	}
	return miss.Grows(path.Join(prefix, kit.Keys(chain)), cache,
		kit.Int(kit.Select("0", strings.TrimPrefix(m.Option(CACHE_OFFEND), "-"))),
		kit.Int(kit.Select("10", m.Option(CACHE_LIMIT))),
		match, value, cb)
}
