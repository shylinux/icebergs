package mdb

import (
	"encoding/json"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func _hash_fields(m *ice.Message) []string {
	return kit.Split(kit.Select(HASH_FIELD, m.OptionFields()))
}
func _hash_select_fields(m *ice.Message, prefix, chain string, key string, field string) (value string) {
	defer m.RLock(prefix, chain)()
	m.Richs(prefix, chain, key, func(h string, v Map) {
		if field == HASH {
			value = h
		} else {
			value = kit.Format(v[field])
		}
	})
	return
}
func _hash_inputs(m *ice.Message, prefix, chain string, field, value string) {
	defer m.RLock(prefix, chain)()

	list := map[string]int{}
	m.Richs(prefix, chain, FOREACH, func(val Map) {
		val = kit.GetMeta(val)
		list[kit.Format(val[field])] += kit.Int(kit.Select("1", val[COUNT]))
	})
	for k, i := range list {
		m.Push(field, k)
		m.Push(COUNT, i)
	}
	m.SortIntR(COUNT)
}
func _hash_insert(m *ice.Message, prefix, chain string, arg ...string) string {
	defer m.Lock(prefix, chain)()

	if value := m.Confm(prefix, kit.Keys(HASH, arg[1])); value != nil {
		value = kit.GetMeta(value)
		for i := 2; i < len(arg)-1; i += 2 {
			kit.Value(value, arg[i], arg[i+1])
		}
		return arg[1]
	}

	m.Log_INSERT(KEY, path.Join(prefix, chain), arg[0], arg[1])
	if expire := m.Conf(prefix, kit.Keys(chain, kit.Keym(EXPIRE))); expire != "" {
		arg = kit.Simple(TIME, m.Time(expire), arg)
	}
	if m.Option(ice.MSG_DOMAIN) != "" {
		m.Conf(prefix, kit.Keys(chain, kit.Keym(SHORT)), m.Conf(prefix, kit.Keym(SHORT)))
	}
	if m.Optionv(TARGET) != nil {
		m.Echo(m.Rich(prefix, chain, kit.Data(arg, TARGET, m.Optionv(TARGET))))
	} else {
		m.Echo(m.Rich(prefix, chain, kit.Data(arg)))
	}
	return m.Result()
}
func _hash_delete(m *ice.Message, prefix, chain, field, value string) {
	defer m.Lock(prefix, chain)()

	if field != HASH {
		field, value = HASH, kit.Select(kit.Hashs(value), m.Option(HASH))
	}
	m.Richs(prefix, chain, value, func(key string, val Map) {
		m.Log_DELETE(KEY, path.Join(prefix, chain), field, value, VALUE, kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, HASH, key), "")
	})
}
func _hash_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	defer m.Lock(prefix, chain)()

	m.Richs(prefix, chain, value, func(key string, val Map) {
		m.Log_MODIFY(KEY, path.Join(prefix, chain), field, value, arg)
		_mdb_modify(m, val, field, arg...)
	})
}
func _hash_select(m *ice.Message, prefix, chain, field, value string) {
	defer m.RLock(prefix, chain)()

	if field == HASH && value == RANDOM {
		value = RANDOMS
	}
	fields := _hash_fields(m)
	m.Richs(prefix, chain, value, func(key string, value Map) {
		_mdb_select(m, key, value, fields, nil)
	})
	if !m.FieldsIsDetail() {
		m.SortTimeR(TIME)
	}
}
func _hash_prunes(m *ice.Message, prefix, chain string, arg ...string) {
	defer m.RLock(prefix, chain)()

	fields := _hash_fields(m)
	m.Richs(prefix, chain, FOREACH, func(key string, val Map) {
		switch val = kit.GetMeta(val); cb := m.OptionCB(PRUNES).(type) {
		case func(string, Map) bool:
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
}
func _hash_export(m *ice.Message, prefix, chain, file string) {
	defer m.Lock(prefix, chain)()

	f, p, e := miss.CreateFile(kit.Keys(file, JSON))
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
	defer m.Lock(prefix, chain)()

	f, e := miss.OpenFile(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	list := Map{}
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

const (
	HASH_FIELD = "time,hash,type,name,text"
)
const HASH = "hash"

func HashAction(args ...Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: AutoConfig(args...),
		INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) { HashInputs(m, arg) }},
		CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) { HashCreate(m, arg) }},
		REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) { HashRemove(m, arg) }},
		MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) { HashModify(m, arg) }},
		SELECT: {Name: "select hash auto", Help: "列表", Hand: func(m *ice.Message, arg ...string) { HashSelect(m, arg...) }},
		PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) { HashPrunes(m, nil) }},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) { HashExport(m, arg) }},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) { HashImport(m, arg) }},
	}
}
func HashActionStatus(args ...Any) ice.Actions {
	list := HashAction(args...)
	list[PRUNES] = &ice.Action{Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
		m.OptionFields(m.Config(FIELD))
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, ice.ERROR)
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, ice.CLOSE)
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, ice.STOP)
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, ice.END)
		m.Tables(func(value ice.Maps) { HashRemove(m, HASH, value[HASH]) })
	}}
	return list
}

func HashShort(m *ice.Message) string {
	return kit.Select(HASH, m.Config(SHORT), m.Config(SHORT) != UNIQ)
}
func HashField(m *ice.Message) string { return kit.Select(HASH_FIELD, m.Config(FIELD)) }
func HashArgs(m *ice.Message, arg ...Any) []string {
	return _mdb_args(m, HashField(m), arg...)
}
func HashInputs(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(INPUTS, m.PrefixKey(), "", HASH, HashArgs(m, arg))
}
func HashCreate(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(INSERT, m.PrefixKey(), "", HASH, HashArgs(m, arg...))
}
func HashRemove(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(DELETE, m.PrefixKey(), "", HASH, m.OptionSimple(HashShort(m)), arg)
}
func HashModify(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmd(MODIFY, m.PrefixKey(), "", HASH, m.OptionSimple(HashShort(m)), HashArgs(m, arg...))
}
func HashSelect(m *ice.Message, arg ...string) *ice.Message {
	m.Fields(len(kit.Slice(arg, 0, 1)), HashField(m))
	m.Cmdy(SELECT, m.PrefixKey(), "", HASH, HashShort(m), arg)
	if m.PushAction(m.Config(ACTION), REMOVE); !m.FieldsIsDetail() {
		return m.StatusTimeCount()
	} else {
		return m.StatusTime()
	}
}
func HashPrunes(m *ice.Message, cb func(Maps) bool) *ice.Message {
	expire := kit.Time(kit.Select(m.Time("-72h"), m.Option(EXPIRE)))
	m.Cmd(m.CommandKey()).Tables(func(value Maps) {
		if cb != nil && !cb(value) || kit.Time(value[TIME]) < expire {
			HashRemove(m, HashShort(m), value[HashShort(m)])
		}
	})
	return m
}
func HashExport(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(EXPORT, m.PrefixKey(), "", HASH, arg)
}
func HashImport(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(IMPORT, m.PrefixKey(), "", HASH, arg)
}

func HashTarget(m *ice.Message, h string, add func() Any) Any {
	defer m.Lock()()

	p := m.Confv(m.PrefixKey(), kit.Keys(HASH, h, TARGET))
	if pp, ok := p.(Map); ok && len(pp) == 0 {
		p = nil
	}

	if p == nil && add != nil {
		p = add()
		m.Confv(m.PrefixKey(), kit.Keys(HASH, h, TARGET), p)
	}
	return p
}
func HashPrunesValue(m *ice.Message, field, value string) {
	m.OptionFields(m.Config(FIELD))
	m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, field, value)
}
func HashSelectFields(m *ice.Message, key string, field string) (value string) {
	HashSelectDetail(m, key, func(h string, val ice.Map) {
		if field == HASH {
			value = h
		} else {
			value = kit.Format(val[field])
		}
	})
	return
}
func HashSelectDetail(m *ice.Message, key string, cb Any) bool {
	defer m.RLock(m.PrefixKey(), "")()
	has := false
	m.Richs(m.PrefixKey(), nil, key, func(key string, value Map) {
		hashSelect(m, key, value, cb)
		has = true
	})
	return has
}
func HashSelectUpdate(m *ice.Message, key string, cb Any) *ice.Message {
	defer m.Lock(m.PrefixKey(), "")()
	m.Richs(m.PrefixKey(), nil, key, func(key string, value Map) {
		hashSelect(m, key, value, cb)
	})
	return m
}
func HashSelectSearch(m *ice.Message, args []string, keys ...string) *ice.Message {
	if len(keys) == 0 {
		ls := kit.Split(m.Config(FIELD))
		for _, k := range ls {
			switch k {
			case TIME, HASH:
			default:
				keys = append(keys, k)
			}
		}
	}
	if args[0] == m.CommandKey() {
		HashSelectValue(m, func(value ice.Map) {
			if args[1] == "" || args[1] == value[keys[1]] {
				m.PushSearch(kit.SimpleKV("", value[keys[0]], value[keys[1]], value[keys[2]]), value)
			}
		})
	}
	return m
}
func HashSelectValue(m *ice.Message, cb Any) *ice.Message {
	defer m.RLock(m.PrefixKey(), "")()
	m.Richs(m.PrefixKey(), nil, FOREACH, func(key string, value Map) {
		hashSelect(m, key, value, cb)
	})
	return m
}
func HashSelects(m *ice.Message, arg ...string) *ice.Message {
	m.OptionFields(m.Config(FIELD))
	return HashSelect(m, arg...)
}
func hashSelect(m *ice.Message, key string, value Map, cb Any) *ice.Message {
	switch value = kit.GetMeta(value); cb := cb.(type) {
	case func(string, Map):
		cb(key, value)
	case func(Map):
		cb(value)
	case func(Any):
		cb(value[TARGET])
	case nil:
	default:
		m.ErrorNotImplement(cb)
	}
	return m
}
