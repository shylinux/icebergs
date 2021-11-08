package mdb

import (
	"encoding/csv"
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _list_fields(m *ice.Message) []string {
	return kit.Split(kit.Select("time,id,type,name,text", m.OptionFields()))
}
func _list_inputs(m *ice.Message, prefix, chain string, field, value string) {
	list := map[string]int{}
	m.Grows(prefix, chain, "", "", func(index int, val map[string]interface{}) {
		if val = kit.GetMeta(val); kit.Format(val[kit.MDB_COUNT]) != "" {
			list[kit.Format(val[field])] = kit.Int(val[kit.MDB_COUNT])
		} else {
			list[kit.Format(val[field])]++
		}
	})
	for k, i := range list {
		m.Push(field, k)
		m.Push(kit.MDB_COUNT, i)
	}
	m.SortIntR(kit.MDB_COUNT)
}
func _list_insert(m *ice.Message, prefix, chain string, arg ...string) {
	m.Log_INSERT(kit.MDB_KEY, path.Join(prefix, chain), arg[0], arg[1])
	m.Echo("%d", m.Grow(prefix, chain, kit.Dict(arg)))
}
func _list_delete(m *ice.Message, prefix, chain, field, value string) {
}
func _list_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Grows(prefix, chain, field, value, func(index int, val map[string]interface{}) {
		val = kit.GetMeta(val)
		m.Log_MODIFY(kit.MDB_KEY, path.Join(prefix, chain), field, value, arg)
		for i := 0; i < len(arg); i += 2 {
			if arg[i] == field {
				continue
			}
			kit.Value(val, arg[i], kit.Select("", arg, i+1))
		}
	})
}
func _list_select(m *ice.Message, prefix, chain, field, value string) {
	if value == "" {
		field = ""
	}
	fields := _list_fields(m)
	m.Grows(prefix, chain, kit.Select(m.Option(ice.CACHE_FIELD), field), kit.Select(m.Option(ice.CACHE_VALUE), value), func(index int, val map[string]interface{}) {
		switch val = kit.GetMeta(val); cb := m.Optionv(kit.Keycb(SELECT)).(type) {
		case func(fields []string, value map[string]interface{}):
			cb(fields, val)
		default:
			if m.OptionFields() == DETAIL {
				m.Push(DETAIL, val)
			} else {
				m.Push("", val, fields)
			}
		}
	})
}
func _list_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := kit.Create(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	count := 0
	head := kit.Split(m.OptionFields())
	m.Grows(prefix, chain, "", "", func(index int, val map[string]interface{}) {
		if val = kit.GetMeta(val); index == 0 {
			if len(head) == 0 || head[0] == "detail" { // 默认表头
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

	m.Log_EXPORT(kit.MDB_KEY, path.Join(prefix, chain), kit.MDB_FILE, p, kit.MDB_COUNT, count)
	m.Conf(prefix, kit.Keys(chain, kit.Keym(kit.MDB_COUNT)), 0)
	m.Conf(prefix, kit.Keys(chain, kit.MDB_LIST), "")
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

const LIST = "list"

func ListAction(fields ...string) map[string]*ice.Action {
	return ice.SelectAction(map[string]*ice.Action{
		INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INPUTS, m.PrefixKey(), "", LIST, arg)
		}},
		INSERT: {Name: "insert type=go name=hi text=hello", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSERT, m.PrefixKey(), "", LIST, arg)
		}},
		DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(DELETE, m.PrefixKey(), "", LIST, m.OptionSimple(kit.MDB_ID), arg)
		}},
		MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(MODIFY, m.PrefixKey(), "", LIST, m.OptionSimple(kit.MDB_ID), arg)
		}},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
			m.Option(ice.CACHE_LIMIT, "-1")
			m.OptionFields(m.Config(kit.MDB_FIELD))
			m.Cmdy(EXPORT, m.PrefixKey(), "", LIST)
			m.Conf(m.PrefixKey(), kit.MDB_LIST, "")
			m.Config(kit.MDB_COUNT, 0)
		}},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(IMPORT, m.PrefixKey(), "", LIST, arg)
		}},
		PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(PRUNES, m.PrefixKey(), "", LIST, arg)
		}},
		PREV: {Name: "prev", Help: "上一页", Hand: func(m *ice.Message, arg ...string) {
			PrevPage(m, m.Config(kit.MDB_COUNT), kit.Slice(arg, 1)...)
		}},
		NEXT: {Name: "next", Help: "下一页", Hand: func(m *ice.Message, arg ...string) {
			NextPage(m, m.Config(kit.MDB_COUNT), kit.Slice(arg, 1)...)
		}},
	}, fields...)
}
func ListSelect(m *ice.Message, arg ...string) *ice.Message {
	m.OptionPage(kit.Slice(arg, 1)...)
	m.Fields(len(kit.Slice(arg, 0, 1)), m.Config(kit.MDB_FIELD))
	m.Cmdy(SELECT, m.PrefixKey(), "", LIST, kit.MDB_ID, arg)
	if !m.FieldsIsDetail() {
		m.StatusTimeCountTotal(m.Config(kit.MDB_COUNT))
	}
	return m
}
