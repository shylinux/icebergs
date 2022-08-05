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
	defer RLock(m, prefix, chain)()

	list := map[string]int{}
	Richs(m, prefix, chain, FOREACH, func(val Map) {
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
	defer Lock(m, prefix, chain)()

	if value := m.Confm(prefix, kit.Keys(HASH, arg[1])); value != nil && arg[1] != "" {
		value = kit.GetMeta(value)
		for i := 2; i < len(arg)-1; i += 2 {
			kit.Value(value, arg[i], arg[i+1])
		}
		return arg[1]
	}

	m.Logs(INSERT, KEY, path.Join(prefix, chain), arg[0], arg[1])
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
		m.Logs(DELETE, KEY, path.Join(prefix, chain), field, value, VALUE, kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, HASH, key), "")
	})
}
func _hash_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	defer Lock(m, prefix, chain)()

	Richs(m, prefix, chain, value, func(key string, val Map) {
		m.Logs(MODIFY, KEY, path.Join(prefix, chain), field, value, arg)
		_mdb_modify(m, val, field, arg...)
	})
}
func _hash_select(m *ice.Message, prefix, chain, field, value string) {
	defer RLock(m, prefix, chain)()

	if field == HASH && value == RANDOM {
		value = RANDOMS
	}
	fields := _hash_fields(m)
	Richs(m, prefix, chain, value, func(key string, value Map) {
		_mdb_select(m, m.OptionCB(""), key, value, fields, nil)
	})
	if !m.FieldsIsDetail() {
		m.SortTimeR(TIME)
	}
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
	Richs(m, prefix, chain, FOREACH, func(key string, val Map) {
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
	defer Lock(m, prefix, chain)()

	f, p, e := miss.CreateFile(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	m.Assert(en.Encode(m.Confv(prefix, kit.Keys(chain, HASH))))

	m.Logs(EXPORT, KEY, path.Join(prefix, chain), FILE, p)
	m.Conf(prefix, kit.Keys(chain, HASH), "")
	m.Echo(p)
}
func _hash_import(m *ice.Message, prefix, chain, file string) {
	defer Lock(m, prefix, chain)()

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
			Rich(m, prefix, chain, data)
			count++
		}
	}

	m.Logs(IMPORT, KEY, path.Join(prefix, chain), COUNT, count)
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
		SELECT: {Name: "select", Help: "列表", Hand: func(m *ice.Message, arg ...string) { HashSelect(m, arg...) }},
		PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) { HashPrunes(m, nil) }},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) { HashExport(m, arg) }},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) { HashImport(m, arg) }},
	}
}
func HashCloseAction(args ...Any) ice.Actions {
	return ice.MergeActions(HashAction(args...), ice.Actions{
		ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { HashSelectClose(m) }},
	})
}
func HashStatusAction(args ...Any) ice.Actions {
	list := HashAction(args...)
	list[PRUNES] = &ice.Action{Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
		m.OptionFields(m.Config(FIELD))
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "error")
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "close")
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "stop")
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "end")
	}}
	return list
}
func HashStatusCloseAction(args ...Any) ice.Actions {
	return ice.MergeActions(HashStatusAction(args...), ice.Actions{
		ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { HashSelectClose(m) }},
	})
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
func HashCreate(m *ice.Message, arg ...Any) string {
	msg := m.Spawn()
	return m.Echo(msg.Cmdx(INSERT, m.PrefixKey(), "", HASH, HashArgs(msg, arg...))).Result()
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
	}
	return m.StatusTime()
}
func HashPrunes(m *ice.Message, cb func(Maps) bool) *ice.Message {
	expire := kit.Time(kit.Select(m.Time("-72h"), m.Option(EXPIRE)))
	m.Cmd("", func(value Maps) {
		if kit.Time(value[TIME]) > expire {
			return
		}
		if cb != nil && !cb(value) {
			return
		}
		HashRemove(m, HashShort(m), value[HashShort(m)])
	})
	return m
}
func HashExport(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(EXPORT, m.PrefixKey(), "", HASH, arg)
}
func HashImport(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(IMPORT, m.PrefixKey(), "", HASH, arg)
}

func HashTarget(m *ice.Message, h string, add func() Any) (p Any) {
	HashSelectUpdate(m, h, func(value ice.Map) {
		p = value[TARGET]
		if pp, ok := p.(Map); ok && len(pp) == 0 {
			p = nil
		}
		if p == nil && add != nil {
			p = add()
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
func HashSelectDetail(m *ice.Message, key string, cb Any) bool {
	defer RLock(m, m.PrefixKey(), "")()
	has := false
	Richs(m, m.PrefixKey(), nil, key, func(key string, value Map) {
		_mdb_select(m, cb, key, value, nil, nil)
		has = true
	})
	return has
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
	HashSelectValue(m, func(target ice.Any) {
		if c, ok := target.(io.Closer); ok {
			m.Logs(DELETE, TARGET, m.PrefixKey())
			c.Close()
		}
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
