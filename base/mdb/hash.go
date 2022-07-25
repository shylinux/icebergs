package mdb

import (
	"encoding/json"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _hash_fields(m *ice.Message) []string {
	return kit.Split(kit.Select("time,hash,type,name,text", m.OptionFields()))
}
func _hash_inputs(m *ice.Message, prefix, chain string, field, value string) {
	defer m.RLock(prefix, chain)()

	list := map[string]int{}
	m.Richs(prefix, chain, FOREACH, func(key string, val ice.Map) {
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
	defer m.Lock(prefix, chain)()

	if m.Option(ice.MSG_DOMAIN) != "" {
		m.Conf(prefix, kit.Keys(chain, kit.Keym(SHORT)), m.Conf(prefix, kit.Keym(SHORT)))
	}
	m.Log_INSERT(KEY, path.Join(prefix, chain), arg[0], arg[1])
	if expire := m.Conf(prefix, kit.Keys(chain, kit.Keym(EXPIRE))); expire != "" {
		arg = kit.Simple(TIME, m.Time(expire), arg)
	}

	m.Echo(m.Rich(prefix, chain, kit.Data(arg)))
}
func _hash_delete(m *ice.Message, prefix, chain, field, value string) {
	defer m.Lock(prefix, chain)()

	if field != HASH {
		field, value = HASH, kit.Select(kit.Hashs(value), m.Option(HASH))
	}
	m.Richs(prefix, chain, value, func(key string, val ice.Map) {
		m.Log_DELETE(KEY, path.Join(prefix, chain), field, value, VALUE, kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, HASH, key), "")
	})
}
func _hash_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	defer m.Lock(prefix, chain)()

	m.Richs(prefix, chain, value, func(key string, val ice.Map) {
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
	defer m.RLock(prefix, chain)()

	if field == HASH && value == RANDOM {
		value = RANDOMS
	}
	fields := _hash_fields(m)
	m.Richs(prefix, chain, value, func(key string, val ice.Map) {
		switch val = kit.GetMeta(val); cb := m.OptionCB(SELECT).(type) {
		case func(fields []string, value ice.Map):
			cb(fields, val)
		case func(value ice.Map):
			cb(val)
		default:
			if m.OptionFields() == DETAIL {
				m.Push(DETAIL, val)
			} else {
				m.Push(key, val, fields)
			}
		}
	})
	if !m.FieldsIsDetail() {
		m.SortTimeR(TIME)
	}
}
func _hash_prunes(m *ice.Message, prefix, chain string, arg ...string) {
	defer m.RLock(prefix, chain)()

	fields := _hash_fields(m)
	m.Richs(prefix, chain, FOREACH, func(key string, val ice.Map) {
		switch val = kit.GetMeta(val); cb := m.OptionCB(PRUNES).(type) {
		case func(string, ice.Map) bool:
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
	defer m.Lock(prefix, chain)()

	f, e := os.Open(kit.Keys(file, JSON))
	if m.Warn(e) {
		return
	}
	defer f.Close()

	list := ice.Map{}
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

const HASH = "hash"

func AutoConfig(args ...ice.Any) *ice.Action {
	return &ice.Action{Hand: func(m *ice.Message, arg ...string) {
		if cs := m.Target().Configs; cs[m.CommandKey()] == nil && len(args) > 0 {
			cs[m.CommandKey()] = &ice.Config{Value: kit.Data(args...)}
			m.Load(m.CommandKey())
		}

		inputs := []ice.Any{}
		kit.Fetch(kit.Split(m.Config(FIELD)), func(i int, k string) {
			switch k {
			case TIME, HASH, ID:
				return
			}
			inputs = append(inputs, k)
		})

		cs := m.Target().Commands
		if cs[m.CommandKey()] == nil {
			return
		}

		if cs[m.CommandKey()].Actions[INSERT] != nil {
			if cs[m.CommandKey()].Meta[INSERT] == nil {
				m.Design(INSERT, "添加", append([]ice.Any{ZONE}, inputs...)...)
			}
		} else if cs[m.CommandKey()].Actions[CREATE] != nil {
			if cs[m.CommandKey()].Meta[CREATE] == nil {
				m.Design(CREATE, "创建", inputs...)
			}
		}
	}}
}
func HashAction(args ...ice.Any) ice.Actions {
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
func HashActionStatus(args ...ice.Any) ice.Actions {
	list := HashAction(args...)
	list[PRUNES] = &ice.Action{Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
		m.OptionFields(m.Config(FIELD))
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "error")
		m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, STATUS, "close")
	}}
	return list
}

func HashInputs(m *ice.Message, arg ...ice.Any) *ice.Message {
	return m.Cmdy(INPUTS, m.PrefixKey(), "", HASH, kit.Simple(arg...))
}
func HashCreate(m *ice.Message, arg ...ice.Any) *ice.Message {
	field := m.Config(FIELD)
	args := kit.Simple(arg...)
	for i := 0; i < len(args); i += 2 {
		if !strings.Contains(field, args[i]) {
			args[i] = kit.Keys("extra", args[i])
		}
	}
	return m.Cmdy(INSERT, m.PrefixKey(), "", HASH, args)
}
func HashRemove(m *ice.Message, arg ...ice.Any) *ice.Message {
	m.OptionFields(m.Config(FIELD))
	defer m.Event(kit.Keys(m.CommandKey(), REMOVE), m.CommandKey(), m.Option(m.Config(SHORT)))
	return m.Cmd(DELETE, m.PrefixKey(), "", HASH, kit.Simple(arg...))
}
func HashModify(m *ice.Message, arg ...ice.Any) *ice.Message {
	field := m.Config(FIELD)
	args := kit.Simple(arg...)
	for i := 0; i < len(args); i += 2 {
		if !strings.Contains(field, args[i]) {
			args[i] = kit.Keys("extra", args[i])
		}
	}
	return m.Cmd(MODIFY, m.PrefixKey(), "", HASH, args)
}
func HashSelect(m *ice.Message, arg ...string) *ice.Message {
	m.Fields(len(arg), m.Config(FIELD))
	m.Cmdy(SELECT, m.PrefixKey(), "", HASH, m.Config(SHORT), arg)
	m.PushAction(m.Config("action"), REMOVE)
	m.StatusTimeCount()
	return m
}
func HashPrunes(m *ice.Message, cb func(ice.Maps) bool) *ice.Message {
	_key := func(m *ice.Message) string {
		if m.Config(HASH) == UNIQ {
			return HASH
		}
		if m.Config(SHORT) == UNIQ {
			return HASH
		}
		return kit.Select(HASH, m.Config(SHORT))
	}
	expire := kit.Time(kit.Select(m.Time("-72h"), m.Option(EXPIRE)))
	m.Cmd(m.CommandKey()).Tables(func(value ice.Maps) {
		if kit.Time(value[TIME]) > expire {
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
func HashExport(m *ice.Message, arg ...ice.Any) *ice.Message {
	m.OptionFields(m.Config(FIELD))
	return m.Cmd(EXPORT, m.PrefixKey(), "", HASH, kit.Simple(arg...))
}
func HashImport(m *ice.Message, arg ...ice.Any) *ice.Message {
	return m.Cmd(IMPORT, m.PrefixKey(), "", HASH, kit.Simple(arg...))
}

func HashCache(m *ice.Message, h string, add func() ice.Any) ice.Any {
	defer m.Lock()()

	p := m.Confv(m.PrefixKey(), kit.Keys(HASH, h, "_cache"))
	if pp, ok := p.(ice.Map); ok && len(pp) == 0 {
		p = nil
	}

	if add != nil && p == nil {
		p = add()
		m.Confv(m.PrefixKey(), kit.Keys(HASH, h, "_cache"), p) // 添加
	}
	return p
}
