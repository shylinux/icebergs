package mdb

import (
	"encoding/csv"
	"os"
	"path"
	"sort"
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _list_fields(m *ice.Message) []string {
	return kit.Split(kit.Select("time,id,type,name,text", strings.Join(kit.Simple(m.Optionv(FIELDS)), ",")))
}
func _list_insert(m *ice.Message, prefix, chain string, arg ...string) {
	m.Log_INSERT(kit.MDB_KEY, path.Join(prefix, chain), arg[0], arg[1])
	m.Echo("%d", m.Grow(prefix, chain, kit.Dict(arg)))
}
func _list_delete(m *ice.Message, prefix, chain, field, value string) {
}
func _list_select(m *ice.Message, prefix, chain, field, value string) {
	if value == "" {
		field = ""
	}
	fields := _list_fields(m)
	cb := m.Optionv(kit.Keycb(SELECT))
	m.Grows(prefix, chain, kit.Select(m.Option(CACHE_FIELD), field), kit.Select(m.Option(CACHE_VALUE), value), func(index int, val map[string]interface{}) {
		val = kit.GetMeta(val)
		switch cb := cb.(type) {
		case func(fields []string, value map[string]interface{}):
			cb(fields, val)
		default:
			if m.Option(FIELDS) == DETAIL {
				m.Push(DETAIL, val)
			} else {
				m.Push("", val, fields)
			}
		}
	})
}
func _list_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Grows(prefix, chain, field, value, func(index int, val map[string]interface{}) {
		val = kit.GetMeta(val)
		for i := 0; i < len(arg); i += 2 {
			if arg[i] == field {
				continue
			}
			kit.Value(val, arg[i], kit.Select("", arg, i+1))
		}
		m.Log_MODIFY(kit.MDB_KEY, path.Join(prefix, chain), field, value, arg)
	})
}
func _list_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := kit.Create(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	count := 0
	head := kit.Split(m.Option(FIELDS))
	m.Grows(prefix, chain, "", "", func(index int, val map[string]interface{}) {
		if val = kit.GetMeta(val); index == 0 {
			if len(head) == 0 { // 默认表头
				for k := range val {
					head = append(head, k)
				}
				sort.Strings(head)
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

	m.Log_EXPORT(kit.MDB_KEY, path.Join(prefix, chain), kit.MDB_FILE, p, kit.MDB_COUNT, count)
	m.Echo(p)
}
func _list_import(m *ice.Message, prefix, chain, file string) {
	f, e := os.Open(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	r := csv.NewReader(f)

	count := 0
	head, _ := r.Read()
	for {
		line, e := r.Read()
		if e != nil {
			break
		}

		data := kit.Dict()
		for i, k := range head {
			if k == kit.MDB_EXTRA {
				kit.Value(data, k, kit.UnMarshal(line[i]))
			} else {
				kit.Value(data, k, line[i])
			}
		}

		m.Grow(prefix, chain, data)
		count++
	}

	m.Log_IMPORT(kit.MDB_KEY, kit.Keys(prefix, chain), kit.MDB_COUNT, count)
	m.Echo("%d", count)
}
func _list_prunes(m *ice.Message, prefix, chain string, arg ...string) {
}
func _list_inputs(m *ice.Message, prefix, chain string, field, value string) {
	list := map[string]int{}
	m.Grows(prefix, chain, "", "", func(index int, val map[string]interface{}) {
		val = kit.GetMeta(val)
		list[kit.Format(val[field])]++
	})
	for k, i := range list {
		m.Push(field, k)
		m.Push(kit.MDB_COUNT, i)
	}
	m.SortIntR(kit.MDB_COUNT)
}
