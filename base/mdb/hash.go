package mdb

import (
	"encoding/json"
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _hash_fields(m *ice.Message) []string {
	return kit.Split(kit.Select("time,hash,type,name,text", m.OptionFields()))
}
func _hash_inputs(m *ice.Message, prefix, chain string, field, value string) {
	list := map[string]int{}
	m.Richs(prefix, chain, FOREACH, func(key string, val map[string]interface{}) {
		if val = kit.GetMeta(val); kit.Format(val[COUNT]) != "" {
			list[kit.Format(val[field])] = kit.Int(val[COUNT])
		} else {
			list[kit.Format(val[field])]++
		}
	})
	for k, i := range list {
		m.Push(field, k)
		m.Push(COUNT, i)
	}
	m.SortIntR(COUNT)
}
func _hash_insert(m *ice.Message, prefix, chain string, arg ...string) {
	if m.Option(ice.MSG_DOMAIN) != "" {
		m.Conf(prefix, kit.Keys(chain, kit.Keym(SHORT)), m.Conf(prefix, kit.Keym(SHORT)))
	}
	m.Log_INSERT(KEY, path.Join(prefix, chain), arg[0], arg[1])
	m.Echo(m.Rich(prefix, chain, kit.Data(arg)))
}
func _hash_delete(m *ice.Message, prefix, chain, field, value string) {
	if field != HASH {
		field, value = HASH, kit.Select(kit.Hashs(value), m.Option(HASH))
	}
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
		m.Log_DELETE(KEY, path.Join(prefix, chain), field, value, VALUE, kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, HASH, key), "")
	})
}
func _hash_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
		val = kit.GetMeta(val)
		m.Log_MODIFY(KEY, path.Join(prefix, chain), field, value, arg)
		for i := 0; i < len(arg); i += 2 {
			if arg[i] == field {
				continue
			}
			kit.Value(val, arg[i], kit.Select("", arg, i+1))
		}
	})
}
func _hash_select(m *ice.Message, prefix, chain, field, value string) {
	if field == HASH && value == RANDOM {
		value = RANDOMS
	}
	fields := _hash_fields(m)
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
		switch val = kit.GetMeta(val); cb := m.Optionv(kit.Keycb(SELECT)).(type) {
		case func(fields []string, value map[string]interface{}):
			cb(fields, val)
		default:
			if m.OptionFields() == DETAIL {
				m.Push(DETAIL, val)
			} else {
				m.Push(key, val, fields)
			}
		}
	})
	if m.Option(FIELDS) != DETAIL {
		m.SortTimeR(TIME)
	}
}
func _hash_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := kit.Create(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	m.Assert(en.Encode(m.Confv(prefix, kit.Keys(chain, HASH))))

	m.Log_EXPORT(KEY, path.Join(prefix, chain), FILE, p)
	m.Conf(prefix, kit.Keys(chain, HASH), "")
	m.Echo(p)
}
func _hash_import(m *ice.Message, prefix, chain, file string) {
	f, e := os.Open(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	list := map[string]interface{}{}
	m.Assert(json.NewDecoder(f).Decode(&list))

	count := 0
	if m.Conf(prefix, kit.Keys(chain, META, SHORT)) == "" {
		for k, data := range list {
			m.Conf(prefix, kit.Keys(chain, HASH, k), data)
			count++
		}
	} else {
		for _, data := range list {
			m.Rich(prefix, chain, data)
			count++
		}
	}

	m.Log_IMPORT(KEY, path.Join(prefix, chain), COUNT, count)
	m.Echo("%d", count)
}
func _hash_prunes(m *ice.Message, prefix, chain string, arg ...string) {
	fields := _hash_fields(m)
	m.Richs(prefix, chain, FOREACH, func(key string, val map[string]interface{}) {
		switch val = kit.GetMeta(val); cb := m.Optionv(kit.Keycb(PRUNES)).(type) {
		case func(string, map[string]interface{}) bool:
			if !cb(key, val) {
				return
			}
		default:
			for i := 0; i < len(arg)-1; i += 2 {
				if val[arg[i]] != arg[i+1] && kit.Value(val, arg[i]) != arg[i+1] {
					return
				}
			}
		}
		m.Push(key, val, fields)
	})
	m.Table(func(index int, value map[string]string, head []string) {
		_hash_delete(m, prefix, chain, HASH, value[HASH])
	})
}

const HASH = "hash"

func HashAction(args ...interface{}) map[string]*ice.Action {
	_key := func(m *ice.Message) string {
		if m.Config(HASH) == "uniq" {
			return HASH
		}
		if m.Config(SHORT) == "uniq" {
			return HASH
		}
		return kit.Select(HASH, m.Config(SHORT))
	}
	return ice.SelectAction(map[string]*ice.Action{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
				cs[m.CommandKey()] = &ice.Config{Value: kit.Data(args...)}
			}
		}},
		INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INPUTS, m.PrefixKey(), "", HASH, arg)
		}},
		CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSERT, m.PrefixKey(), "", HASH, arg)
		}},
		REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
			m.OptionFields(m.Config(FIELD))
			m.Cmdy(DELETE, m.PrefixKey(), "", HASH, m.OptionSimple(_key(m)), arg)
			m.Event(kit.Keys(m.CommandKey(), REMOVE), m.CommandKey(), m.Option(m.Config(SHORT)))
		}},
		MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(MODIFY, m.PrefixKey(), "", HASH, m.OptionSimple(_key(m)), arg)
		}},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
			m.OptionFields(m.Config(kit.META_FIELD))
			m.Cmdy(EXPORT, m.PrefixKey(), "", HASH, arg)
		}},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(IMPORT, m.PrefixKey(), "", HASH, arg)
		}},
		PRUNES: &ice.Action{Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
			HashPrunes(m, nil)
		}},
		SELECT: &ice.Action{Name: "select hash auto", Help: "列表", Hand: func(m *ice.Message, arg ...string) {
			HashSelect(m, arg...)
		}},
	})
}
func HashActionStatus(args ...interface{}) map[string]*ice.Action {
	list := HashAction(args...)
	list[PRUNES] = &ice.Action{Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
		m.OptionFields(m.Config(FIELD))
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "error")
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "close")
	}}
	return list
}
func HashSelect(m *ice.Message, arg ...string) *ice.Message {
	m.Fields(len(arg), m.Config(FIELD))
	m.Cmdy(SELECT, m.PrefixKey(), "", HASH, m.Config(SHORT), arg)
	m.PushAction(REMOVE)
	m.StatusTimeCount()
	return m
}
func HashPrunes(m *ice.Message, cb func(map[string]string) bool) *ice.Message {
	_key := func(m *ice.Message) string {
		if m.Config(HASH) == "uniq" {
			return HASH
		}
		return kit.Select(HASH, m.Config(SHORT))
	}
	before := kit.Time(kit.Select(m.Time("-72h"), m.Option(kit.MDB_BEFORE)))
	m.Cmd(m.CommandKey()).Table(func(index int, value map[string]string, head []string) {
		if kit.Time(value[TIME]) > before {
			return
		}
		if cb != nil && cb(value) {
			return
		}
		m.OptionFields(m.Config(FIELD))
		m.Cmdy(DELETE, m.PrefixKey(), "", HASH, _key(m), value[_key(m)])
	})
	return m
}
