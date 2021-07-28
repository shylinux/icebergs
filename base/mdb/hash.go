package mdb

import (
	"encoding/json"
	"os"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _hash_fields(m *ice.Message) []string {
	return kit.Split(kit.Select("time,hash,type,name,text", strings.Join(kit.Simple(m.Optionv(FIELDS)), ",")))
}
func _hash_insert(m *ice.Message, prefix, chain string, arg ...string) {
	if m.Option(ice.MSG_DOMAIN) != "" {
		m.Conf(prefix, kit.Keys(chain, kit.MDB_META, kit.MDB_SHORT), m.Conf(prefix, kit.Keys(kit.MDB_META, kit.MDB_SHORT)))
	}
	m.Log_INSERT(kit.MDB_KEY, path.Join(prefix, chain), arg[0], arg[1])
	h := m.Rich(prefix, chain, kit.Data(arg))
	m.Echo(h)
}
func _hash_delete(m *ice.Message, prefix, chain, field, value string) {
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
		m.Log_DELETE(kit.MDB_KEY, path.Join(prefix, chain), field, value, kit.MDB_VALUE, kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, kit.MDB_HASH, key), "")
	})
}
func _hash_select(m *ice.Message, prefix, chain, field, value string) {
	if field == kit.MDB_HASH && value == RANDOM {
		value = kit.MDB_RANDOMS
	}
	fields := _hash_fields(m)
	cb := m.Optionv(kit.Keycb(SELECT))
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
		val = kit.GetMeta(val)
		switch cb := cb.(type) {
		case func(fields []string, value map[string]interface{}):
			cb(fields, val)
		default:
			if m.Option(FIELDS) == DETAIL {
				m.Push(DETAIL, val)
			} else {
				m.Push(key, val, fields)
			}
		}
	})
	if m.Option(FIELDS) != DETAIL {
		m.SortTimeR(kit.MDB_TIME)
	}
}
func _hash_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
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
func _hash_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := kit.Create(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	e = en.Encode(m.Confv(prefix, kit.Keys(chain, HASH)))

	m.Log_EXPORT(kit.MDB_KEY, path.Join(prefix, chain), kit.MDB_FILE, p)
	m.Echo(p)
}
func _hash_import(m *ice.Message, prefix, chain, file string) {
	f, e := os.Open(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	list := map[string]interface{}{}
	de := json.NewDecoder(f)
	de.Decode(&list)

	count := 0
	if m.Conf(prefix, kit.Keys(chain, kit.MDB_META, kit.MDB_SHORT)) == "" {
		for k, data := range list {
			m.Conf(prefix, kit.Keys(chain, kit.MDB_HASH, k), data)
			count++
		}
	} else {
		for _, data := range list {
			m.Rich(prefix, chain, data)
			count++
		}
	}

	m.Log_IMPORT(kit.MDB_KEY, path.Join(prefix, chain), kit.MDB_COUNT, count)
	m.Echo("%d", count)
}
func _hash_prunes(m *ice.Message, prefix, chain string, arg ...string) {
	fields := _hash_fields(m)
	m.Richs(prefix, chain, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}
		switch cb := m.Optionv(kit.Keycb(PRUNES)).(type) {
		case func(string, map[string]interface{}) bool:
			if !cb(key, val) {
				return
			}
		default:
			for i := 0; i < len(arg)-1; i += 2 {
				if val[arg[i]] != arg[i+1] {
					return
				}
			}
		}
		m.Push(key, val, fields)
	})
	m.Table(func(index int, value map[string]string, head []string) {
		_hash_delete(m, prefix, chain, kit.MDB_HASH, value[kit.MDB_HASH])
	})
}
func _hash_inputs(m *ice.Message, prefix, chain string, field, value string) {
	list := map[string]int{}
	m.Richs(prefix, chain, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		val = kit.GetMeta(val)
		if field == kit.MDB_HASH {
			list[key]++
		} else {
			list[kit.Format(val[field])]++
		}
	})
	for k, i := range list {
		m.Push(field, k)
		m.Push(kit.MDB_COUNT, i)
	}
	m.Sort(kit.MDB_COUNT, "int_r")
}

func HashAction(key string, fields ...string) map[string]*ice.Action {
	list := map[string]*ice.Action{
		MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(MODIFY, m.Prefix(key), "", HASH, m.OptionSimple(kit.MDB_HASH), arg)
		}},
		REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(DELETE, m.Prefix(key), "", HASH, m.OptionSimple(kit.MDB_HASH))
		}},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(EXPORT, m.Prefix(key), "", HASH)
		}},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(IMPORT, m.Prefix(key), "", HASH)
		}},
		INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INPUTS, m.Prefix(key), "", HASH, arg)
		}},
	}
	if len(fields) == 0 {
		return list
	}

	res := map[string]*ice.Action{}
	for _, field := range fields {
		res[field] = list[field]
	}
	return res
}

const HASH = "hash"
