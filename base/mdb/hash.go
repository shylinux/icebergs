package mdb

import (
	"encoding/json"
	"io"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
	"shylinux.com/x/toolkits/miss"
)

func _hash_fields(m *ice.Message) []string {
	return kit.Split(kit.Select(HASH_FIELD, m.OptionFields()))
}
func _hash_inputs(m *ice.Message, prefix, chain string, field, value string) {
	list := map[string]int{}
	defer func() {
		delete(list, "")
		for k, i := range list {
			m.Push(field, k).Push(COUNT, i)
		}
		m.SortIntR(COUNT)
	}()
	defer RLock(m, prefix, chain)()
	Richs(m, prefix, chain, FOREACH, func(key string, value Map) {
		value = kit.GetMeta(value)
		list[kit.Format(value[field])] += kit.Int(kit.Select("1", value[COUNT]))
	})
}
func _hash_insert(m *ice.Message, prefix, chain string, arg ...string) string {
	m.Logs(INSERT, KEY, path.Join(prefix, chain), arg)
	defer Lock(m, prefix, chain)()
	if expire := m.Conf(prefix, kit.Keys(chain, kit.Keym(EXPIRE))); expire != "" {
		arg = kit.Simple(TIME, m.Time(expire), arg)
	}
	return m.Echo(Rich(m, prefix, chain, kit.Data(arg, TARGET, m.Optionv(TARGET)))).Result()
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
	fields := _hash_fields(m)
	defer RLock(m, prefix, chain)()
	Richs(m, prefix, chain, value, func(key string, value Map) { _mdb_select(m, m.OptionCB(""), key, value, fields, nil) })
}
func _hash_select_field(m *ice.Message, prefix, chain string, key string, field string) (value string) {
	defer RLock(m, prefix, chain)()
	Richs(m, prefix, chain, key, func(key string, val Map) {
		if field == HASH {
			value = key
		} else {
			value = kit.Format(val[field])
		}
	})
	return
}
func _hash_prunes(m *ice.Message, prefix, chain string, arg ...string) {
	fields := _hash_fields(m)
	defer RLock(m, prefix, chain)()
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
	defer Lock(m, prefix, chain)()
	f, p, e := miss.CreateFile(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()
	defer m.Echo(p)
	m.Logs(EXPORT, KEY, path.Join(prefix, chain), FILE, p, COUNT, len(m.Confm(prefix, kit.Keys(chain, HASH))))
	en := json.NewEncoder(f)
	if en.SetIndent("", "  "); !m.Warn(en.Encode(m.Confv(prefix, kit.Keys(chain, HASH))), EXPORT, prefix) {
		m.Conf(prefix, kit.Keys(chain, HASH), "")
	}
}
func _hash_import(m *ice.Message, prefix, chain, file string) {
	defer Lock(m, prefix, chain)()
	f, e := miss.OpenFile(kit.Keys(file, JSON))
	if m.Warn(e) {
		return
	}
	defer f.Close()
	list := Map{}
	m.Assert(json.NewDecoder(f).Decode(&list))
	m.Logs(IMPORT, KEY, path.Join(prefix, chain), FILE, kit.Keys(file, JSON), COUNT, len(list))
	defer m.Echo("%d", len(list))
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
	return ice.Actions{
		ice.CTX_INIT: AutoConfig(append(kit.List(FIELD, HASH_FIELD), arg...)...),
		ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { HashSelectClose(m) }},

		INPUTS: {Hand: func(m *ice.Message, arg ...string) { HashInputs(m, arg) }},
		CREATE: {Hand: func(m *ice.Message, arg ...string) { HashCreate(m, arg) }},
		REMOVE: {Hand: func(m *ice.Message, arg ...string) { HashRemove(m, arg) }},
		MODIFY: {Hand: func(m *ice.Message, arg ...string) { HashModify(m, arg) }},
		SELECT: {Name: "select hash auto create", Hand: func(m *ice.Message, arg ...string) { HashSelect(m, arg...) }},
		PRUNES: {Name: "prunes before@date", Hand: func(m *ice.Message, arg ...string) { HashPrunes(m, nil) }},
		EXPORT: {Hand: func(m *ice.Message, arg ...string) { HashExport(m, arg) }},
		IMPORT: {Hand: func(m *ice.Message, arg ...string) { HashImport(m, arg) }},
	}
}
func StatusHashAction(arg ...Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		PRUNES: &ice.Action{Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "error", STATUS, "close", STATUS, "stop", STATUS, "end", ice.OptionFields(HashField(m)))
		}},
	}, HashAction(arg...))
}
func ClearHashOnExitAction() ice.Actions {
	return ice.MergeActions(ice.Actions{ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { m.Conf("", HASH, "") }}})
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
func HashInputs(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(INPUTS, m.PrefixKey(), "", HASH, arg)
}
func HashCreate(m *ice.Message, arg ...Any) string {
	if len(arg) == 0 {
		arg = append(arg, m.OptionSimple(HashField(m)))
	}
	return m.Echo(m.Cmdx(append(kit.List(INSERT, m.PrefixKey(), "", HASH, logs.FileLineMeta(-1)), arg...)...)).Result()
}
func HashRemove(m *ice.Message, arg ...Any) *ice.Message {
	if args := kit.Simple(arg); len(args) == 0 {
		arg = append(arg, m.OptionSimple(HashKey(m)))
	} else if len(args) == 1 {
		arg = kit.List(HashKey(m), args[0])
	}
	return m.Cmdy(DELETE, m.PrefixKey(), "", HASH, arg)
}
func HashModify(m *ice.Message, arg ...Any) *ice.Message {
	if args := kit.Simple(arg...); args[0] != HASH && args[0] != HashShort(m) {
		arg = append(kit.List(m.OptionSimple(HashKey(m))), arg...)
	}
	return m.Cmd(MODIFY, m.PrefixKey(), "", HASH, arg)
}
func HashSelect(m *ice.Message, arg ...string) *ice.Message {
	if len(arg) > 0 && arg[0] == FOREACH {
		m.Fields(0, HashField(m))
	} else {
		m.Fields(len(kit.Slice(arg, 0, 1)), HashField(m))
	}
	m.Cmdy(SELECT, m.PrefixKey(), "", HASH, HashShort(m), arg, logs.FileLineMeta(-1))
	if m.PushAction(m.Config(ACTION), REMOVE); !m.FieldsIsDetail() {
		return m.StatusTimeCount()
	}
	return m.StatusTime()
}
func HashPrunes(m *ice.Message, cb func(Map) bool) *ice.Message {
	expire := kit.Select(m.Time("-"+kit.Select("72h", m.Config(EXPIRE))), m.Option("before"))
	m.OptionCB(PRUNES, func(key string, value Map) bool {
		if kit.Format(value[TIME]) > expire {
			return false
		}
		return cb == nil || cb(value)
	})
	return m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, ice.OptionFields(HashField(m))).StatusTimeCount()
}
func HashExport(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(EXPORT, m.PrefixKey(), "", HASH, arg)
}
func HashImport(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(IMPORT, m.PrefixKey(), "", HASH, arg)
}

func HashSelects(m *ice.Message, arg ...string) *ice.Message {
	m.OptionFields(HashField(m))
	return HashSelect(m, arg...)
}
func HashSelectValue(m *ice.Message, cb Any) *ice.Message {
	m.OptionFields(m.Config(FIELD))
	defer RLock(m, m.PrefixKey(), "")()
	Richs(m, m.PrefixKey(), nil, FOREACH, func(key string, value Map) { _mdb_select(m, cb, key, value, nil, nil) })
	return m
}
func HashSelectUpdate(m *ice.Message, key string, cb Any) *ice.Message {
	defer Lock(m, m.PrefixKey(), "")()
	Richs(m, m.PrefixKey(), nil, key, func(key string, value Map) { _mdb_select(m, cb, key, value, nil, nil) })
	return m
}
func HashSelectDetail(m *ice.Message, key string, cb Any) (has bool) {
	defer RLock(m, m.PrefixKey(), "")()
	Richs(m, m.PrefixKey(), nil, key, func(key string, value Map) {
		_mdb_select(m, cb, key, value, nil, nil)
		has = true
	})
	return
}
func HashSelectDetails(m *ice.Message, key string, cb func(ice.Map) bool) ice.Map {
	val := kit.Dict()
	HashSelectDetail(m, key, func(value ice.Map) {
		if cb(value) {
			for k, v := range value {
				val[k] = v
			}
		}
	})
	return val
}
func HashSelectField(m *ice.Message, key string, field string) (value string) {
	HashSelectDetail(m, key, func(key string, val ice.Map) {
		if field == HASH {
			value = key
		} else {
			value = kit.Format(val[field])
		}
	})
	return
}
func HashSelectTarget(m *ice.Message, key string, create Any) (target Any) {
	HashSelectUpdate(m, key, func(value ice.Map) {
		target = value[TARGET]
		if _target, ok := target.([]string); ok && len(_target) == 0 {
			target = nil
		}
		if _target, ok := target.(ice.List); ok && len(_target) == 0 {
			target = nil
		}
		if _target, ok := target.(Map); ok && len(_target) == 0 {
			target = nil
		}
		if target != nil || create == nil {
			return
		}
		switch create := create.(type) {
		case func(ice.Map) ice.Any:
			target = create(value)
		case func() ice.Any:
			target = create()
		default:
			m.ErrorNotImplement(create)
			return
		}
		value[TARGET] = target
	})
	return
}
func HashSelectClose(m *ice.Message) *ice.Message {
	HashSelectValue(m, func(value ice.Map) {
		if c, ok := value[TARGET].(io.Closer); ok {
			c.Close()
		}
		delete(value, TARGET)
	})
	return m
}
func HashPrunesValue(m *ice.Message, field, value string) {
	m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, field, value, ice.OptionFields(HashField(m)))
}
func HashCreateDeferRemove(m *ice.Message, arg ...Any) func() {
	h := HashCreate(m, arg...)
	return func() { HashRemove(m.SetResult(), HASH, h) }
}
func HashModifyDeferRemove(m *ice.Message, arg ...Any) func() {
	HashModify(m, arg...)
	return func() { HashRemove(m, arg...) }
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
