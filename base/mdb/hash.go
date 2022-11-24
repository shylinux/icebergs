package mdb

import (
	"encoding/json"
	"io"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func _hash_fields(m *ice.Message) []string {
	return kit.Split(kit.Select(HASH_FIELD, m.OptionFields()))
}
func _hash_inputs(m *ice.Message, prefix, chain string, field, value string) {
	switch field {
	case EXPIRE:
		m.Push(field, "72h")
		m.Push(field, "24h")
		m.Push(field, "8h")
		m.Push(field, "3h")
		m.Push(field, "1h")
		return
	}
	defer RLock(m, prefix, chain)()
	list := map[string]int{}
	Richs(m, prefix, chain, FOREACH, func(key string, val Map) {
		val = kit.GetMeta(val)
		list[kit.Format(val[field])] += kit.Int(kit.Select("1", val[COUNT]))
	})
	for k, i := range list {
		if k != "" {
			m.Push(field, k).Push(COUNT, i)
		}
	}
	m.SortIntR(COUNT)
}
func _hash_insert(m *ice.Message, prefix, chain string, arg ...string) string {
	m.Logs(INSERT, KEY, path.Join(prefix, chain), arg[0], arg[1])
	defer Lock(m, prefix, chain)()
	if value := kit.GetMeta(m.Confm(prefix, kit.Keys(HASH, arg[1]))); value != nil && arg[1] != "" {
		kit.Fetch(arg[2:], func(k, v string) { kit.Value(value, k, v)})
		return arg[1]
	}
	if expire := m.Conf(prefix, kit.Keys(chain, kit.Keym(EXPIRE))); expire != "" {
		arg = kit.Simple(TIME, m.Time(expire), arg)
	}
	if m.Optionv(TARGET) != nil && m.Option(TARGET) != "" {
		m.Echo(Rich(m, prefix, chain, kit.Data(arg, TARGET, m.Optionv(TARGET))))
	} else {
		m.Echo(Rich(m, prefix, chain, kit.Data(arg)))
	}
	return m.Result()
}
func _hash_delete(m *ice.Message, prefix, chain, field, value string) {
	defer Lock(m, prefix, chain)()
	Richs(m, prefix, chain, value, func(key string, val Map) {
		if target, ok := kit.GetMeta(val)[TARGET].(io.Closer); ok {
			target.Close()
		}
		m.Logs(DELETE, KEY, path.Join(prefix, chain), field, value, VALUE, kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, HASH, key), "")
	})
}
func _hash_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Logs(MODIFY, KEY, path.Join(prefix, chain), field, value, arg)
	defer Lock(m, prefix, chain)()
	Richs(m, prefix, chain, value, func(key string, val Map) { _mdb_modify(m, val, field, arg...) })
}
func _hash_select(m *ice.Message, prefix, chain, field, value string) {
	if field == HASH && value == RANDOM {
		value = RANDOMS
	}
	defer m.SortTimeR(TIME)
	defer RLock(m, prefix, chain)()
	fields := _hash_fields(m)
	Richs(m, prefix, chain, value, func(key string, value Map) { _mdb_select(m, m.OptionCB(""), key, value, fields, nil) })
}
func _hash_select_field(m *ice.Message, prefix, chain string, key string, field string) (value string) {
	defer RLock(m, prefix, chain)()
	Richs(m, prefix, chain, key, func(h string, v Map) {
		if field == HASH {
			value = h
		} else {
			value = kit.Format(v[field])
		}
	})
	return
}
func _hash_prunes(m *ice.Message, prefix, chain string, arg ...string) {
	defer RLock(m, prefix, chain)()
	fields := _hash_fields(m)
	Richs(m, prefix, chain, FOREACH, func(key string, value Map) {
		switch value = kit.GetMeta(value); cb := m.OptionCB(PRUNES).(type) {
		case func(string, Map) bool:
			if cb(key, value) {
				m.Push(key, value, fields)
			}
		default:
			kit.Fetch(arg, func(k, v string) {
				if value[k] == v || kit.Value(value, k) == v {
					m.Push(key, value, fields)
				}
			})
		}
	})
}
func _hash_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := miss.CreateFile(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()
	m.Logs(EXPORT, KEY, path.Join(prefix, chain), FILE, p)
	defer m.Echo(p).StatusTime(LINK, "/share/local/"+p)
	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	defer Lock(m, prefix, chain)()
	m.Warn(en.Encode(m.Confv(prefix, kit.Keys(chain, HASH))), EXPORT, p)
	m.Conf(prefix, kit.Keys(chain, HASH), "")
}
func _hash_import(m *ice.Message, prefix, chain, file string) {
	f, e := miss.OpenFile(kit.Keys(file, JSON))
	if m.Warn(e) {
		return
	}
	defer f.Close()
	list := Map{}
	m.Assert(json.NewDecoder(f).Decode(&list))
	m.Logs(IMPORT, KEY, path.Join(prefix, chain), COUNT, len(list))
	defer m.Echo("%d", len(list))
	defer Lock(m, prefix, chain)()
	for k, data := range list {
		if m.Confv(prefix, kit.Keys(chain, HASH, k)) == nil {
			m.Confv(prefix, kit.Keys(chain, HASH, k), data)
		} else {
			Rich(m, prefix, chain, data)
		}
	}
}

const (
	HASH_FIELD = "time,hash,type,name,text"
)
const HASH = "hash"

func HashAction(arg ...Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: AutoConfig(arg...), ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { HashSelectClose(m) }},
		INPUTS: {Hand: func(m *ice.Message, arg ...string) { HashInputs(m, arg) }},
		CREATE: {Hand: func(m *ice.Message, arg ...string) { HashCreate(m, arg) }},
		REMOVE: {Hand: func(m *ice.Message, arg ...string) { HashRemove(m, arg) }},
		MODIFY: {Hand: func(m *ice.Message, arg ...string) { HashModify(m, arg) }},
		SELECT: {Hand: func(m *ice.Message, arg ...string) { HashSelect(m, arg...) }},
		PRUNES: {Name: "prunes before@date", Hand: func(m *ice.Message, arg ...string) { HashPrunes(m, nil) }},
		EXPORT: {Hand: func(m *ice.Message, arg ...string) { HashExport(m, arg) }},
		IMPORT: {Hand: func(m *ice.Message, arg ...string) { HashImport(m, arg) }},
	}
}
func HashStatusAction(arg ...Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		PRUNES: &ice.Action{Hand: func(m *ice.Message, arg ...string) {
			m.OptionFields(m.Config(FIELD))
			m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "error")
			m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "close")
			m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "stop")
			m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "end")
		}},
	}, HashAction(arg...))
}
func HashCloseAction(args ...Any) ice.Actions {
	return ice.MergeActions(HashAction(args...), ice.Actions{
		ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { HashSelectClose(m) }},
	})
}
func HashStatusCloseAction(args ...Any) ice.Actions {
	return ice.MergeActions(HashStatusAction(args...), ice.Actions{
		ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { HashSelectClose(m) }},
	})
}

func HashKey(m *ice.Message) string {
	if m.Option(HASH) != "" {
		return HASH
	}
	return HashShort(m)
}
func HashShort(m *ice.Message) string {
	return kit.Select(HASH, m.Config(SHORT), m.Config(SHORT) != UNIQ)
}
func HashField(m *ice.Message) string { return kit.Select(HASH_FIELD, m.Config(FIELD)) }
func HashArgs(m *ice.Message, arg ...Any) []string {
	return _mdb_args(m, "", arg...)
}
func HashInputs(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(INPUTS, m.PrefixKey(), "", HASH, HashArgs(m, arg))
}
func HashCreate(m *ice.Message, arg ...Any) string {
	msg := m.Spawn()
	args := HashArgs(msg, arg...)
	if len(args) == 0 {
		args = m.OptionSimple(m.Config(FIELD))
	}
	return m.Echo(msg.Cmdx(INSERT, m.PrefixKey(), "", HASH, args)).Result()
}
func HashRemove(m *ice.Message, arg ...Any) *ice.Message {
	args := kit.Simple(arg)
	if len(args) == 0 {
		args = m.OptionSimple(HashKey(m))
	} else if len(args) == 1 {
		args = []string{HashKey(m), args[0]}
	}
	return m.Cmdy(DELETE, m.PrefixKey(), "", HASH, args)
}
func HashModify(m *ice.Message, arg ...Any) *ice.Message {
	args := HashArgs(m, arg...)
	if args[0] != HashShort(m) && args[0] != HASH {
		args = append(m.OptionSimple(HashKey(m)), args...)
	}
	return m.Cmd(MODIFY, m.PrefixKey(), "", HASH, args)
}
func HashSelect(m *ice.Message, arg ...string) *ice.Message {
	if len(arg) > 0 && arg[0] == FOREACH {
		m.Fields(0, HashField(m))
	} else {
		m.Fields(len(kit.Slice(arg, 0, 1)), HashField(m))
	}
	m.Cmdy(SELECT, m.PrefixKey(), "", HASH, HashShort(m), arg)
	if m.PushAction(m.Config(ACTION), REMOVE); !m.FieldsIsDetail() {
		return m.StatusTimeCount()
	}
	return m.StatusTime()
}
func HashPrunes(m *ice.Message, cb func(Maps) bool) *ice.Message {
	expire := kit.Select(m.Time("-72h"), m.Option("before"))
	m.Cmd("", func(value Maps) {
		if value[TIME] > expire {
			return
		}
		if cb != nil && !cb(value) {
			return
		}
		HashRemove(m, HashShort(m), value[HashShort(m)])
	})
	return m.StatusTimeCount()
}
func HashExport(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(EXPORT, m.PrefixKey(), "", HASH, arg)
}
func HashImport(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(IMPORT, m.PrefixKey(), "", HASH, arg)
}

func HashTarget(m *ice.Message, h string, add Any) (p Any) {
	HashSelectUpdate(m, h, func(value ice.Map) {
		p = value[TARGET]
		if pp, ok := p.(Map); ok && len(pp) == 0 {
			p = nil
		}
		if p == nil && add != nil {
			switch add := add.(type) {
			case func(ice.Map) ice.Any:
				p = add(value)
			case func() ice.Any:
				p = add()
			default:
				m.ErrorNotImplement(p)
				return
			}
			value[TARGET] = p
		}
	})
	return
}
func HashPrunesValue(m *ice.Message, field, value string) {
	m.OptionFields(m.Config(FIELD))
	m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, field, value)
}
func HashSelectField(m *ice.Message, key string, field string) (value string) {
	HashSelectDetail(m, key, func(h string, val ice.Map) {
		if field == HASH {
			value = h
		} else {
			value = kit.Format(val[field])
		}
	})
	return
}
func HashSelectDetail(m *ice.Message, key string, cb Any) (has bool) {
	defer RLock(m, m.PrefixKey(), "")()
	Richs(m, m.PrefixKey(), nil, key, func(key string, value Map) {
		_mdb_select(m, cb, key, value, nil, nil)
		has = true
	})
	return
}
func HashSelectUpdate(m *ice.Message, key string, cb Any) *ice.Message {
	defer Lock(m, m.PrefixKey(), "")()
	Richs(m, m.PrefixKey(), nil, key, func(key string, value Map) {
		_mdb_select(m, cb, key, value, nil, nil)
	})
	return m
}
func HashSelectValue(m *ice.Message, cb Any) *ice.Message {
	defer RLock(m, m.PrefixKey(), "")()
	Richs(m, m.PrefixKey(), nil, FOREACH, func(key string, value Map) {
		_mdb_select(m, cb, key, value, nil, nil)
	})
	return m
}
func HashSelectClose(m *ice.Message) *ice.Message {
	HashSelectValue(m, func(value ice.Map) {
		target := value[TARGET]
		if c, ok := target.(io.Closer); ok {
			c.Close()
		}
		delete(value, TARGET)
	})
	return m
}
func HashSelects(m *ice.Message, arg ...string) *ice.Message {
	m.OptionFields(m.Config(FIELD))
	return HashSelect(m, arg...)
}

func Richs(m *ice.Message, prefix string, chain Any, raw Any, cb Any) (res Map) {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		return nil
	}
	return miss.Richs(path.Join(prefix, kit.Keys(chain)), cache, raw, cb)

}
func Rich(m *ice.Message, prefix string, chain Any, data Any) string {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		cache = kit.Data()
		m.Confv(prefix, chain, cache)
	}
	if m.Option(SHORT) != "" {
		kit.Value(cache, kit.Keym(SHORT), m.Option(SHORT))
	}
	return miss.Rich(path.Join(prefix, kit.Keys(chain)), cache, data)
}
