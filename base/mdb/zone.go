package mdb

import (
	"encoding/csv"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
	"shylinux.com/x/toolkits/miss"
)

func _zone_fields(m *ice.Message) []string {
	return kit.Split(kit.Select(ZONE_FIELD, m.OptionFields()))
}
func _zone_inputs(m *ice.Message, prefix, chain, zone string, field, value string) {
	if field == m.Conf(prefix, kit.Keys(chain, kit.Keym(SHORT))) {
		_hash_inputs(m, prefix, chain, field, value)
		return
	}
	h := _hash_select_field(m, prefix, chain, zone, HASH)
	defer RLock(m, prefix, chain)()
	_list_inputs(m, prefix, kit.Keys(chain, HASH, h), field, value)
}
func _zone_insert(m *ice.Message, prefix, chain, zone string, arg ...string) {
	h := _hash_select_field(m, prefix, chain, zone, HASH)
	if h == "" {
		h = _hash_insert(m, prefix, chain, m.Conf(prefix, kit.Keys(chain, kit.Keym(SHORT))), zone)
	}
	m.Assert(h != "")
	defer Lock(m, prefix, chain)()
	_list_insert(m, prefix, kit.Keys(chain, HASH, h), arg...)
}
func _zone_modify(m *ice.Message, prefix, chain, zone, id string, arg ...string) {
	h := _hash_select_field(m, prefix, chain, zone, HASH)
	m.Assert(h != "")
	defer RLock(m, prefix, chain)()
	_list_modify(m, prefix, kit.Keys(chain, HASH, h), ID, id, arg...)
}
func _zone_select(m *ice.Message, prefix, chain, zone string, id string) {
	if zone == "" {
		_hash_select(m, prefix, chain, HASH, FOREACH)
		return
	}
	if zone == RANDOM {
		zone = RANDOMS
	}

	fields := _zone_fields(m)
	defer RLock(m, prefix, chain)()
	Richs(m, prefix, chain, kit.Select(FOREACH, zone), func(key string, val Map) {
		chain := kit.Keys(chain, HASH, key)
		defer RLock(m, prefix, chain)()
		Grows(m, prefix, chain, ID, id, func(value ice.Map) {
			_mdb_select(m, m.OptionCB(""), key, value, fields, val)
		})
	})
}
func _zone_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := miss.CreateFile(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	fields := _zone_fields(m)
	if kit.IndexOf(fields, EXTRA) == -1 {
		fields = append(fields, EXTRA)
	}
	w.Write(fields)

	keys := []string{}
	Richs(m, prefix, chain, FOREACH, func(key string, val ice.Map) { keys = append(keys, key) })
	kit.Sort(keys)

	count := 0
	defer Lock(m, prefix, chain)()
	for _, key := range keys {
		Richs(m, prefix, chain, key, func(key string, val ice.Map) {
			val = kit.GetMeta(val)

			chain := kit.Keys(chain, HASH, key)
			defer RLock(m, prefix, chain)()
			Grows(m, prefix, chain, "", "", func(value ice.Map) {
				value = kit.GetMeta(value)

				list := []string{}
				for _, k := range fields {
					list = append(list, kit.Select(kit.Format(kit.Value(val, k)), kit.Format(kit.Value(value, k))))
				}
				w.Write(list)
				count++
			})
		})
	}

	m.Logs(EXPORT, KEY, path.Join(prefix, chain), FILE, p, COUNT, count)
	m.Conf(prefix, kit.Keys(chain, HASH), "")
	m.Echo(p)
}
func _zone_import(m *ice.Message, prefix, chain, file string) {
	f, e := miss.OpenFile(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	r := csv.NewReader(f)
	head, _ := r.Read()
	count := 0

	list := ice.Maps{}
	zkey := kit.Select(head[0], m.OptionFields())

	defer Lock(m, prefix, chain)()
	for {
		line, e := r.Read()
		if e != nil {
			break
		}

		zone := ""
		data := kit.Dict()
		for i, k := range head {
			switch k {
			case zkey:
				zone = line[i]
			case ID:
				continue
			case EXTRA:
				kit.Value(data, k, kit.UnMarshal(line[i]))
			default:
				kit.Value(data, k, line[i])
			}
		}
		if list[zone] == "" {
			list[zone] = Rich(m, prefix, chain, kit.Data(zkey, zone))
		}

		func() {
			chain := kit.Keys(chain, HASH, list[zone])
			defer Lock(m, prefix, chain)()
			Grow(m, prefix, chain, data)
		}()
		count++
	}

	m.Logs(IMPORT, KEY, path.Join(prefix, chain), COUNT, count)
	m.Echo("%d", count)
}

const (
	ZONE_FIELD = "zone,id,time,type,name,text"
)
const ZONE = "zone"

func ZoneAction(args ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: AutoConfig(args...),
		INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) { ZoneInputs(m, arg) }},
		CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) { ZoneCreate(m, arg) }},
		REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) { ZoneRemove(m, arg) }},
		INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) { ZoneInsert(m, arg) }},
		MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) { ZoneModify(m, arg) }},
		SELECT: {Name: "select", Help: "列表", Hand: func(m *ice.Message, arg ...string) { ZoneSelect(m, arg...) }},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) { ZoneExport(m, arg) }},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) { ZoneImport(m, arg) }},
		PREV:   {Name: "prev", Help: "上一页", Hand: func(m *ice.Message, arg ...string) { PrevPage(m, arg[0], arg[1:]...) }},
		NEXT:   {Name: "next", Help: "下一页", Hand: func(m *ice.Message, arg ...string) { NextPageLimit(m, arg[0], arg[1:]...) }},
	}
}
func ZoneShort(m *ice.Message) string {
	return kit.Select(ZONE, m.Config(SHORT), m.Config(SHORT) != UNIQ)
}
func ZoneField(m *ice.Message) string {
	return kit.Select(ZONE_FIELD, m.Config(FIELD))
}
func ZoneArgs(m *ice.Message, arg ...Any) []string {
	return _mdb_args(m, ZoneField(m), arg...)
}

func ZoneInputs(m *ice.Message, arg ...Any) {
	m.Cmdy(INPUTS, m.PrefixKey(), "", ZONE, m.Option(ZoneShort(m)), ZoneArgs(m, arg...))
}
func ZoneCreate(m *ice.Message, arg ...Any) {
	m.Cmdy(INSERT, m.PrefixKey(), "", HASH, arg)
}
func ZoneRemove(m *ice.Message, arg ...Any) {
	args := kit.Simple(arg...)
	if len(args) == 0 {
		args = m.OptionSimple(ZoneShort(m), HASH)
	} else if len(args) == 1 {
		args = []string{ZoneShort(m), args[0]}
	}
	m.Cmdy(DELETE, m.PrefixKey(), "", HASH, args)
}
func ZoneInsert(m *ice.Message, arg ...Any) {
	args := kit.Simple(arg...)
	if len(args) == 0 {
		args = m.OptionSimple(ZoneShort(m), m.Config(FIELD))
	}
	for i := len(args) - 2; i >= 0; i -= 2 {
		if args[i+1] == "" {
			args = args[:i]
		} else {
			break
		}
	}
	m.Cmdy(INSERT, m.PrefixKey(), "", ZONE, args[1], ZoneArgs(m, args[2:]))
}
func ZoneModify(m *ice.Message, arg ...Any) {
	m.Cmdy(MODIFY, m.PrefixKey(), "", ZONE, m.Option(ZoneShort(m)), m.Option(ID), ZoneArgs(m, arg...))
}
func ZoneSelect(m *ice.Message, arg ...string) *ice.Message {
	arg = kit.Slice(arg, 0, 2)
	m.Fields(len(arg), kit.Fields(TIME, m.Config(SHORT), COUNT), m.Config(FIELD))
	if m.Cmdy(SELECT, m.PrefixKey(), "", ZONE, arg, logs.FileLineMeta(logs.FileLine(-1, 3))); len(arg) == 0 {
		if m.Config(SHORT) != "" {
			m.Sort(m.Config(SHORT))
		}
		m.PushAction(m.Config(ACTION), REMOVE)
		m.StatusTimeCount()
	} else if len(arg) == 1 {
		m.StatusTimeCountTotal(m.Conf("", kit.Keys(HASH, HashSelectField(m, arg[0], HASH), kit.Keym(COUNT))))
	}
	return m
}
func ZoneExport(m *ice.Message, arg ...Any) {
	m.Debug("what %v", m.OptionFields())
	if m.OptionFields() == "" {
		m.OptionFields(m.Config(SHORT), m.Config(FIELD))	
	}
	m.Debug("what %v", m.OptionFields())
	m.Cmdy(EXPORT, m.PrefixKey(), "", ZONE, arg)
}
func ZoneImport(m *ice.Message, arg ...Any) {
	m.Cmdy(IMPORT, m.PrefixKey(), "", ZONE, arg)
}
func ZoneSelectPage(m *ice.Message, arg ...string) *ice.Message {
	OptionPages(m, kit.Slice(arg, 2)...)
	return ZoneSelect(m, arg...)
}
func ZoneSelects(m *ice.Message, arg ...string) *ice.Message {
	m.OptionFields(m.Config(FIELD))
	return ZoneSelect(m, arg...)
}
func ZoneSelectAll(m *ice.Message, arg ...string) *ice.Message {
	m.Option(CACHE_LIMIT, "-1")
	return ZoneSelect(m, arg...)
}
func ZoneSelectCB(m *ice.Message, zone string, cb Any) *ice.Message {
	m.OptionCB(SELECT, cb)
	m.Option(CACHE_LIMIT, "-1")
	return ZoneSelect(m, zone)
}
