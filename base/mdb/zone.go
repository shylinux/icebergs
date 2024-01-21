package mdb

import (
	"encoding/csv"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
	"shylinux.com/x/toolkits/miss"
)

func _zone_meta(m *ice.Message, prefix, chain, key string) string {
	defer RLock(m, prefix)()
	return m.Conf(prefix, kit.Keys(chain, kit.Keym(key)))
}
func _zone_fields(m *ice.Message) []string {
	return kit.Split(kit.Select(ZONE_FIELD, m.OptionFields()))
}
func _zone_inputs(m *ice.Message, prefix, chain, zone string, field, value string) {
	if field == _zone_meta(m, prefix, chain, SHORT) {
		_hash_inputs(m, prefix, chain, field, value)
		return
	}
	h := _hash_select_field(m, prefix, chain, zone, HASH)
	_list_inputs(m, prefix, kit.Keys(chain, HASH, h), field, value)
}
func _zone_insert(m *ice.Message, prefix, chain, zone string, arg ...string) {
	h := _hash_select_field(m, prefix, chain, zone, HASH)
	if h == "" {
		h = _hash_insert(m, prefix, chain, kit.Select(ZONE, _zone_meta(m, prefix, chain, SHORT)), zone)
	}
	m.Assert(h != "")
	_list_insert(m, prefix, kit.Keys(chain, HASH, h), arg...)
}
func _zone_modify(m *ice.Message, prefix, chain, zone, id string, arg ...string) {
	h := _hash_select_field(m, prefix, chain, zone, HASH)
	m.Assert(h != "")
	_list_modify(m, prefix, kit.Keys(chain, HASH, h), ID, id, arg...)
}
func _zone_select(m *ice.Message, prefix, chain, zone string, id string) {
	if zone == "" {
		_hash_select(m, prefix, chain, HASH, FOREACH)
		return
	} else if zone == RANDOM {
		zone = RANDOMS
	}
	defer m.SortIntR(ID)
	fields := _zone_fields(m)
	defer RLock(m, prefix)()
	Richs(m, prefix, chain, kit.Select(FOREACH, zone), func(key string, val Map) {
		chain := kit.Keys(chain, HASH, key)
		Grows(m, prefix, chain, ID, id, func(value ice.Map) {
			_mdb_select(m, m.OptionCB(""), key, value, fields, val)
		})
	})
}
func _zone_export(m *ice.Message, prefix, chain, file string) {
	defer Lock(m, prefix)()
	f, p, e := miss.CreateFile(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()
	defer m.Echo(p)
	w := csv.NewWriter(f)
	defer w.Flush()
	head := kit.AddUniq(_zone_fields(m), EXTRA)
	w.Write(head)
	count := 0
	for _, key := range kit.SortedKey(m.Confv(prefix, kit.Keys(chain, HASH))) {
		Richs(m, prefix, chain, key, func(key string, val ice.Map) {
			val = kit.GetMeta(val)
			chain := kit.Keys(chain, HASH, key)
			Grows(m, prefix, chain, "", "", func(value ice.Map) {
				value = kit.GetMeta(value)
				w.Write(kit.Simple(head, func(k string) string {
					return kit.Select(kit.Format(kit.Value(val, k)), kit.Format(kit.Value(value, k)))
				}))
				count++
			})
		})
		m.Conf(prefix, kit.Keys(chain, HASH, key, LIST), "")
		m.Conf(prefix, kit.Keys(chain, HASH, key, META, COUNT), "")
	}
	if count == 0 {
		os.Remove(p)
		return
	}
	m.Logs(EXPORT, KEY, path.Join(prefix, chain), FILE, p, COUNT, count)
}
func _zone_import(m *ice.Message, prefix, chain, file string) {
	defer Lock(m, prefix)()
	f, e := ice.Info.Open(m, kit.Keys(file, CSV))
	if os.IsNotExist(e) {
		return
	}
	if e != nil {
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	head, _ := r.Read()
	zkey := kit.Select(head[0], m.OptionFields())
	list := ice.Maps{}
	times := ice.Maps{}
	kit.For(m.Confv(prefix, kit.Keys(chain, HASH)), func(key string, value ice.Any) {
		times[key] = kit.Format(kit.Value(value, kit.Keys(META, TIME)))
	})
	count := 0
	for {
		line, e := r.Read()
		if e != nil {
			break
		}
		zone, data := "", kit.Dict()
		for i, k := range head {
			switch k {
			case zkey:
				zone = line[i]
			case ID:
				continue
			case EXTRA:
				if line[i] != "" {
					kit.Value(data, k, kit.UnMarshal(line[i]))
				}
			default:
				kit.Value(data, k, line[i])
			}
		}
		if list[zone] == "" {
			list[zone] = Rich(m, prefix, chain, kit.Data(zkey, zone))
			kit.If(times[list[zone]], func(t string) { m.Confv(prefix, kit.Keys(chain, HASH, list[zone], META, TIME), t) })
		}
		func() {
			chain := kit.Keys(chain, HASH, list[zone])
			Grow(m, prefix, chain, data)
		}()
		count++
	}
	m.Logs(IMPORT, KEY, path.Join(prefix, chain), FILE, kit.Keys(file, CSV), COUNT, count)
	m.Echo("%d", count)
}

const (
	ZONE_FIELD = "time,id,type,name,text"
)
const ZONE = "zone"

func ZoneConfig(arg ...Any) *ice.Action {
	return &ice.Action{Hand: func(m *ice.Message, args ...string) {
		if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
			cs[m.CommandKey()] = &ice.Config{Value: kit.Data(arg...)}
		} else {
			kit.For(kit.Dict(arg...), func(k string, v Any) { Config(m, k, v) })
		}
		if cmd := m.Target().Commands[m.CommandKey()]; cmd == nil {
			return
		} else {
			s := kit.Select(ZONE, Config(m, SHORT))
			kit.If(s == UNIQ || strings.Contains(s, ","), func() { s = HASH })
			if cmd.Name == "" {
				cmd.Name = kit.Format("%s %s id auto", m.CommandKey(), s)
				cmd.List = ice.SplitCmd(cmd.Name, cmd.Actions)
			}
			add := func(list []string) (inputs []Any) {
				kit.For(list, func(k string) {
					kit.If(!kit.IsIn(k, TIME, HASH, COUNT, ID), func() {
						inputs = append(inputs, k+kit.Select("", "*", strings.Contains(s, k)))
					})
				})
				return
			}
			kit.If(cmd.Meta[INSERT] == nil, func() { m.Design(INSERT, "", add(kit.Simple(kit.Split(s), kit.Split(ZoneField(m))))...) })
			kit.If(cmd.Meta[CREATE] == nil, func() { m.Design(CREATE, "", add(kit.Split(kit.Select(s, Config(m, FIELD))))...) })
		}
	}}
}
func ZoneAction(arg ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: ZoneConfig(append(kit.List(SHORT, ZONE, FIELDS, ZONE_FIELD), arg...)...),
		INPUTS: {Hand: func(m *ice.Message, arg ...string) { ZoneInputs(m, arg) }},
		CREATE: {Hand: func(m *ice.Message, arg ...string) { ZoneCreate(m, arg) }},
		REMOVE: {Hand: func(m *ice.Message, arg ...string) { ZoneRemove(m, arg) }},
		INSERT: {Hand: func(m *ice.Message, arg ...string) { ZoneInsert(m, arg) }},
		MODIFY: {Hand: func(m *ice.Message, arg ...string) { ZoneModify(m, arg) }},
		SELECT: {Hand: func(m *ice.Message, arg ...string) { ZoneSelect(m, arg...) }},
		EXPORT: {Hand: func(m *ice.Message, arg ...string) { ZoneExport(m, arg) }},
		IMPORT: {Hand: func(m *ice.Message, arg ...string) { ZoneImport(m, arg) }},
	}
}
func ExportZoneAction(arg ...ice.Any) ice.Actions {
	return ice.MergeActions(ZoneAction(arg...), ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			Config(m, IMPORTANT, ice.TRUE)
			ZoneImport(m, arg)
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { m.OptionFields(""); ZoneExport(m, arg) }},
	})
}
func PageZoneAction(arg ...ice.Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		SELECT: {Hand: func(m *ice.Message, arg ...string) { PageZoneSelect(m, arg...) }},
		PREV:   {Hand: func(m *ice.Message, arg ...string) { PrevPageLimit(m, arg[0], arg[1:]...) }},
		NEXT:   {Hand: func(m *ice.Message, arg ...string) { NextPage(m, arg[0], arg[1:]...) }},
	}, ZoneAction(arg...))
}
func ZoneKey(m *ice.Message) string {
	if m.Option(HASH) != "" {
		return HASH
	}
	return ZoneShort(m)
}
func ZoneShort(m *ice.Message) string {
	return kit.Select(ZONE, Config(m, SHORT), Config(m, SHORT) != UNIQ)
}
func ZoneField(m *ice.Message) string { return kit.Select(ZONE_FIELD, Config(m, FIELDS)) }
func ZoneInputs(m *ice.Message, arg ...Any) {
	m.Cmdy(INPUTS, m.PrefixKey(), "", ZONE, m.Option(ZoneKey(m)), arg)
}
func ZoneCreate(m *ice.Message, arg ...Any) { m.Cmdy(INSERT, m.PrefixKey(), "", HASH, arg) }
func ZoneRemove(m *ice.Message, arg ...Any) {
	if args := kit.Simple(arg...); len(args) == 0 {
		arg = append(arg, m.OptionSimple(ZoneKey(m)))
	} else if len(args) == 1 {
		arg = kit.List(ZoneKey(m), args[0])
	}
	m.Cmdy(DELETE, m.PrefixKey(), "", HASH, arg)
}
func ZoneInsert(m *ice.Message, arg ...Any) {
	if args := kit.Simple(arg...); len(args) == 0 {
		m.Cmdy(INSERT, m.PrefixKey(), "", ZONE, m.Option(ZoneShort(m)), m.OptionSimple(ZoneField(m)))
	} else if args[0] == ZoneKey(m) {
		m.Cmdy(INSERT, m.PrefixKey(), "", ZONE, args[1:])
	} else {
		m.Cmdy(INSERT, m.PrefixKey(), "", ZONE, arg)
	}
}
func ZoneModify(m *ice.Message, arg ...Any) {
	if args := kit.Simple(arg...); m.Option(ID) == "" {
		HashModify(m, arg...)
	} else if args[0] == HASH || args[0] == ZoneShort(m) {
		m.Cmdy(MODIFY, m.PrefixKey(), "", ZONE, args[1], args[3], arg[4:])
	} else {
		m.Cmdy(MODIFY, m.PrefixKey(), "", ZONE, m.Option(ZoneKey(m)), m.Option(ID), arg)
	}
}
func ZoneSelect(m *ice.Message, arg ...string) *ice.Message {
	arg = kit.Slice(arg, 0, 2)
	m.Fields(len(arg), kit.Select(kit.Fields(TIME, Config(m, SHORT), COUNT), Config(m, FIELD)), ZoneField(m))
	if m.Cmdy(SELECT, m.PrefixKey(), "", ZONE, arg, logs.FileLineMeta(-1)); len(arg) == 0 {
		m.Sort(ZoneShort(m)).PushAction(Config(m, ACTION), REMOVE).Action(CREATE)
	} else if len(arg) == 1 {
		m.Action(INSERT).StatusTimeCountTotal(_zone_meta(m, m.PrefixKey(), kit.Keys(HASH, HashSelectField(m, arg[0], HASH)), COUNT), "step", "0")
	}
	return m
}
func ZoneExport(m *ice.Message, arg ...Any) {
	kit.If(m.OptionFields() == "", func() { m.OptionFields(Config(m, SHORT), ZoneField(m)) })
	m.Cmdy(EXPORT, m.PrefixKey(), "", ZONE, arg)
}
func ZoneImport(m *ice.Message, arg ...Any) {
	m.Cmdy(IMPORT, m.PrefixKey(), "", ZONE, arg)
}
func ZoneSelects(m *ice.Message, arg ...string) *ice.Message {
	m.OptionFields(ZoneField(m))
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
func PageZoneSelect(m *ice.Message, arg ...string) *ice.Message {
	OptionPages(m, kit.Slice(arg, 2)...)
	arg = kit.Slice(arg, 0, 2)
	if ZoneSelect(m, arg...); len(arg) == 0 {
		m.Action(CREATE)
	} else if len(arg) == 1 {
		m.Action(INSERT, "page")
	}
	return m
}
