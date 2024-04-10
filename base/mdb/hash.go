package mdb

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web/html"
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
		kit.For(list, func(k string, i int) { m.Push(field, k).Push(COUNT, i) })
		m.SortIntR(COUNT)
	}()
	defer RLock(m, prefix)()
	Richs(m, prefix, chain, FOREACH, func(key string, value Map) {
		value = kit.GetMeta(value)
		list[kit.Format(value[field])] += kit.Int(kit.Select("1", value[COUNT]))
	})
}
func _hash_insert(m *ice.Message, prefix, chain string, arg ...string) string {
	m.Logs(INSERT, KEY, path.Join(prefix, chain), arg)
	defer Lock(m, prefix)()
	if expire := m.Conf(prefix, kit.Keys(chain, kit.Keym(EXPIRE))); expire != "" && arg[0] != HASH {
		arg = kit.Simple(TIME, m.Time(expire), arg)
	}
	if arg[0] == HASH {
		m.Echo(arg[1]).Conf(prefix, kit.Keys(chain, HASH, arg[1]), kit.Data(arg[2:]))
	} else {
		if target, ok := m.Optionv(TARGET).([]string); ok && len(target) == 0 {
			m.Echo(Rich(m, prefix, chain, kit.Data(arg)))
		} else {
			m.Echo(Rich(m, prefix, chain, kit.Data(arg, TARGET, m.Optionv(TARGET))))
		}
	}
	saveImportant(m, prefix, chain, kit.Simple(INSERT, prefix, chain, HASH, HASH, m.Result(), TIME, m.Time(), arg)...)
	return m.Result()
}
func _hash_delete(m *ice.Message, prefix, chain, field, value string) {
	defer Lock(m, prefix)()
	Richs(m, prefix, chain, value, func(key string, val Map) {
		if target, ok := kit.GetMeta(val)[TARGET].(io.Closer); ok {
			target.Close()
		}
		m.Logs(DELETE, KEY, path.Join(prefix, chain), field, value, VALUE, kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, HASH, key), "")
		saveImportant(m, prefix, chain, kit.Simple(DELETE, prefix, chain, HASH, HASH, key)...)
	})
}
func _hash_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Logs(MODIFY, KEY, path.Join(prefix, chain), field, value, arg)
	defer Lock(m, prefix)()
	Richs(m, prefix, chain, value, func(key string, val Map) {
		_mdb_modify(m, val, field, arg...)
		saveImportant(m, prefix, chain, kit.Simple(MODIFY, prefix, chain, HASH, HASH, key, arg)...)
	})
}
func _hash_select(m *ice.Message, prefix, chain, field, value string) {
	kit.If(field == HASH && value == RANDOM, func() { value = RANDOMS })
	defer m.SortStrR(TIME)
	fields := _hash_fields(m)
	defer RLock(m, prefix)()
	if strings.Contains(value, ",") {
		kit.For(kit.Split(value), func(value string) {
			Richs(m, prefix, chain, value, func(key string, value Map) { _mdb_select(m, m.OptionCB(""), key, value, fields, nil) })
		})
	} else {
		Richs(m, prefix, chain, value, func(key string, value Map) { _mdb_select(m, m.OptionCB(""), key, value, fields, nil) })
	}
}
func _hash_select_field(m *ice.Message, prefix, chain string, key string, field string) (value string) {
	defer RLock(m, prefix)()
	Richs(m, prefix, chain, key, func(key string, val Map) { value = kit.Select(kit.Format(val[field]), key, field == HASH) })
	return
}
func _hash_prunes(m *ice.Message, prefix, chain string, arg ...string) {
	fields := _hash_fields(m)
	kit.If(kit.IndexOf(fields, HASH) == -1, func() { fields = append(fields, HASH) })
	defer RLock(m, prefix)()
	Richs(m, prefix, chain, FOREACH, func(key string, value Map) {
		switch value = kit.GetMeta(value); cb := m.OptionCB("").(type) {
		case func(string, Map) bool:
			kit.If(cb(key, value), func() { m.Push(key, value, fields) })
		default:
			kit.For(arg, func(k, v string) {
				kit.If(value[k] == v || kit.Value(value, k) == v, func() { m.Push(key, value, fields) })
			})
		}
	})
}
func _hash_export(m *ice.Message, prefix, chain, file string) {
	defer Lock(m, prefix)()
	p := kit.Keys(file, JSON)
	count := len(Confm(m, prefix, kit.Keys(chain, HASH)))
	if count == 0 {
		if s, e := os.Stat(p); e == nil && !s.IsDir() {
			os.Remove(p)
		}
		return
	}
	f, p, e := miss.CreateFile(p)
	m.Assert(e)
	defer f.Close()
	defer m.Echo(p)
	m.Logs(EXPORT, KEY, path.Join(prefix, chain), FILE, p, COUNT, count)
	en := json.NewEncoder(f)
	if en.SetIndent("", "  "); !m.WarnNotValid(en.Encode(m.Confv(prefix, kit.Keys(chain, HASH))), EXPORT, prefix) {
		m.Conf(prefix, kit.Keys(chain, HASH), "")
	}
}
func _hash_import(m *ice.Message, prefix, chain, file string) {
	defer Lock(m, prefix)()
	f, e := miss.OpenFile(kit.Keys(file, JSON))
	if e != nil && !ice.Info.Important {
		return
	} else if m.WarnNotFound(e) {
		return
	}
	defer f.Close()
	data := Map{}
	m.Assert(json.NewDecoder(f).Decode(&data))
	m.Logs(IMPORT, KEY, path.Join(prefix, chain), FILE, kit.Keys(file, JSON), COUNT, len(data))
	kit.If(m.Confv(prefix, kit.Keys(chain, HASH)) == nil, func() { m.Confv(prefix, kit.Keys(chain, HASH), ice.Map{}) })
	kit.For(data, func(k string, v Any) { m.Confv(prefix, kit.Keys(chain, HASH, k), v) })
	m.Echo("%d", len(data))
}

const (
	MONTH = "720h"
	DAYS  = "72h"
	HOUR  = "1h"

	CACHE_CLEAR_ONEXIT = "cache.clear.onexit"
)
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
		SELECT: {Hand: func(m *ice.Message, arg ...string) { HashSelect(m, arg...) }},
		PRUNES: {Name: "prunes before@date", Hand: func(m *ice.Message, arg ...string) { HashPrunes(m, nil) }},
		EXPORT: {Hand: func(m *ice.Message, arg ...string) { HashExport(m, arg) }},
		IMPORT: {Hand: func(m *ice.Message, arg ...string) { HashImport(m, arg) }},
	}
}
func StatusHashAction(arg ...Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		PRUNES: &ice.Action{Name: "prunes status", Hand: func(m *ice.Message, arg ...string) {
			args := []string{}
			kit.For(kit.Split(m.OptionDefault(STATUS, "error,close,stop,end")), func(s string) { args = append(args, STATUS, s) })
			m.Cmdy(PRUNES, m.PrefixKey(), m.Option(SUBKEY), HASH, args, ice.OptionFields(HashField(m)))
		}},
	}, HashAction(arg...))
}
func ClearOnExitHashAction() ice.Actions {
	return ice.Actions{ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { Conf(m, m.PrefixKey(), HASH, "") }}}
}
func ExportHashAction(arg ...Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { Config(m, IMPORTANT, ice.TRUE); HashImport(m, arg) }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { HashExport(m, arg) }},
	}, HashAction(arg...))
}

const (
	DEV_REQUEST  = "devRequest"
	DEV_CHOOSE   = "devChoose"
	DEV_RESPONSE = "devResponse"
	DEV_CONFIRM  = "devConfirm"
	DEV_CREATE   = "devCreate"
)

func DevDataAction(fields ...string) ice.Actions {
	const (
		DAEMON = "daemon"
		ORIGIN = "origin"
		BACK   = "back"
	)
	return ice.Actions{
		DEV_REQUEST: {Name: "request origin*", Help: "请求", Icon: "bi bi-cloud-download", Hand: func(m *ice.Message, arg ...string) {
			back := m.Options(ice.MSG_USERWEB, m.Option(ice.MSG_USERHOST)).MergePod("")
			m.ProcessOpen(m.Options(ice.MSG_USERWEB, m.Option(ORIGIN), ice.MSG_USERPOD, "").MergePodCmd("", m.PrefixKey(), ACTION, DEV_CHOOSE, BACK, back, DAEMON, m.Option(ice.MSG_DAEMON)))
		}},
		DEV_CHOOSE: {Hand: func(m *ice.Message, arg ...string) {
			HashSelect(m.Options(ice.MSG_FIELDS, kit.Join(fields))).PushAction(DEV_RESPONSE).Options(ice.MSG_ACTION, "")
		}},
		DEV_RESPONSE: {Help: "选择", Hand: func(m *ice.Message, arg ...string) {
			if !m.WarnNotAllow(m.Option(ice.MSG_METHOD) != http.MethodPost) {
				m.ProcessReplace(m.ParseLink(m.Option(BACK)).MergePodCmd("", m.PrefixKey(), ACTION, DEV_CONFIRM, m.OptionSimple(DAEMON), m.OptionSimple(fields...)))
			}
		}},
		DEV_CONFIRM: {Hand: func(m *ice.Message, arg ...string) {
			m.EchoInfoButton(kit.JoinWord(m.PrefixKey(), m.Cmdx("nfs.cat", "src/template/mdb.hash/savefrom.html"), m.Option(kit.Split(fields[0])[0])), DEV_CREATE)
		}},
		DEV_CREATE: {Help: "创建", Hand: func(m *ice.Message, arg ...string) {
			if !m.WarnNotAllow(m.Option(ice.MSG_METHOD) != http.MethodPost) {
				defer kit.If(m.Option(DAEMON), func(p string) { m.Cmd("space", p, "refresh") })
				HashCreate(m.ProcessClose(), m.OptionSimple(fields...))
			}
		}},
	}
}

func HashKey(m *ice.Message) string {
	if m.Option(HASH) != "" {
		return HASH
	}
	return HashShort(m)
}
func HashShort(m *ice.Message) string {
	if m.Option(SHORT) != "" {
		return m.Option(SHORT)
	}
	short := ""
	if m.Option(SUBKEY) != "" {
		if short = Conf(m, m.PrefixKey(), kit.Keys(m.Option(SUBKEY), META, SHORT)); short == "" {
			short = Config(m, SHORTS)
		}
	} else {
		short = Config(m, SHORT)
	}
	return kit.Select(HASH, short, short != UNIQ)
}
func HashField(m *ice.Message) string {
	if m.Option(FIELD) != "" {
		return m.Option(FIELD)
	}
	field := ""
	if m.Option(SUBKEY) != "" {
		if field = Conf(m, m.PrefixKey(), kit.Keys(m.Option(SUBKEY), META, FIELDS)); field == "" {
			field = Config(m, FIELDS)
		}
	} else {
		field = Config(m, FIELD)
	}
	return kit.Select(HASH_FIELD, field)
}
func HashInputs(m *ice.Message, arg ...Any) *ice.Message {
	return m.Cmdy(INPUTS, m.PrefixKey(), m.Option(SUBKEY), HASH, arg)
}
func HashCreate(m *ice.Message, arg ...Any) string {
	kit.If(len(arg) == 0 || len(kit.Simple(arg...)) == 0, func() {
		arg = append(arg, m.OptionSimple(kit.Filters(kit.Split(HashField(m)), TIME, HASH)...))
	})
	kit.If(m.Option(SUBKEY) == "", func() { kit.If(Config(m, SHORTS), func(p string) { arg = append([]ice.Any{SHORT, p}, arg) }) })
	return m.Echo(m.Cmdx(append(kit.List(INSERT, m.PrefixKey(), m.Option(SUBKEY), HASH, logs.FileLineMeta(-1)), arg...)...)).Result()
}
func HashRemove(m *ice.Message, arg ...Any) *ice.Message {
	if args := kit.Simple(arg...); len(args) == 0 {
		arg = append(arg, m.OptionSimple(HashKey(m)))
	} else if len(args) == 1 {
		arg = kit.List(HashKey(m), args[0])
	}
	return m.Cmdy(DELETE, m.PrefixKey(), m.Option(SUBKEY), HASH, arg)
}
func HashModify(m *ice.Message, arg ...Any) *ice.Message {
	if args := kit.Simple(arg...); args[0] != HASH && args[0] != HashShort(m) {
		arg = append(kit.List(m.OptionSimple(HashKey(m))), arg...)
	}
	return m.Cmd(MODIFY, m.PrefixKey(), m.Option(SUBKEY), HASH, arg)
}
func HashSelect(m *ice.Message, arg ...string) *ice.Message {
	if len(arg) > 0 && (arg[0] == FOREACH || strings.Contains(arg[0], ",")) {
		m.Fields(0, HashField(m))
	} else {
		m.Fields(len(kit.Slice(arg, 0, 1)), HashField(m))
	}
	m.Cmdy(SELECT, m.PrefixKey(), m.Option(SUBKEY), HASH, HashShort(m), arg, logs.FileLineMeta(-1))
	kit.If(kit.Select(Config(m, SHORT), Config(m, SORT)), func(sort string) { kit.If(sort != UNIQ, func() { m.Sort(sort) }) })
	if m.PushAction(Config(m, ACTION), REMOVE); !m.FieldsIsDetail() {
		m.Options(ice.TABLE_CHECKBOX, Config(m, html.CHECKBOX))
		return m.Action(CREATE, PRUNES)
	}
	return sortByField(m, HashField(m), arg...)
}
func HashPrunes(m *ice.Message, cb func(Map) bool) *ice.Message {
	expire := kit.Select(m.Time("-"+kit.Select(DAYS, Config(m, EXPIRE))), m.Option("before"))
	m.OptionCB(PRUNES, func(key string, value Map) bool { return kit.Format(value[TIME]) < expire && (cb == nil || cb(value)) })
	return m.Cmdy(PRUNES, m.PrefixKey(), "", HASH, ice.OptionFields(HashField(m)))
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
	m.OptionFields(Config(m, FIELD))
	defer RLock(m, m.PrefixKey())()
	Richs(m, m.PrefixKey(), nil, FOREACH, func(key string, value Map) { _mdb_select(m, cb, key, value, nil, nil) })
	return m
}
func HashSelectUpdate(m *ice.Message, key string, cb Any) *ice.Message {
	defer Lock(m, m.PrefixKey())()
	Richs(m, m.PrefixKey(), nil, kit.Select(FOREACH, key), func(key string, value Map) { _mdb_select(m, cb, key, value, nil, nil) })
	return m
}
func HashSelectDetail(m *ice.Message, key string, cb Any) (has bool) {
	defer RLock(m, m.PrefixKey())()
	Richs(m, m.PrefixKey(), nil, key, func(key string, value Map) { _mdb_select(m, cb, key, value, nil, nil); has = true })
	return
}
func HashSelectDetails(m *ice.Message, key string, cb func(Map) bool) Map {
	val := kit.Dict()
	HashSelectDetail(m, key, func(value Map) { kit.If(cb(value), func() { kit.For(value, func(k string, v Any) { val[k] = v }) }) })
	return val
}
func HashSelectField(m *ice.Message, key string, field string) (value string) {
	HashSelectDetail(m, key, func(key string, val Map) { value = kit.Select(kit.Format(kit.Value(val, field)), key, field == HASH) })
	return
}
func HashSelectTarget(m *ice.Message, key string, create Any) (target Any) {
	HashSelectUpdate(m, key, func(value Map) {
		target = value[TARGET]
		if _target, ok := target.([]string); ok && len(_target) == 0 {
			target = nil
		}
		if _target, ok := target.(List); ok && len(_target) == 0 {
			target = nil
		}
		if _target, ok := target.(Map); ok && len(_target) == 0 {
			target = nil
		}
		if target != nil || create == nil {
			return
		}
		switch create := create.(type) {
		case func(Maps) Any:
			target = create(kit.ToMaps(value))
		case func(Map) Any:
			target = create(value)
		case func() Any:
			target = create()
		default:
			m.ErrorNotImplement(create)
		}
		value[TARGET] = target
	})
	return
}
func HashSelectClose(m *ice.Message) *ice.Message {
	HashSelectValue(m, func(value Map) {
		if c, ok := value[TARGET].(io.Closer); ok {
			m.WarnNotValid(c.Close())
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
	cache := Confm(m, prefix, chain)
	if cache == nil {
		return nil
	}
	if value := kit.Format(raw); strings.Contains(value, ",") {
		kit.For(kit.Split(value), func(value string) {
			res = miss.Richs(path.Join(prefix, kit.Keys(chain)), cache, value, cb)
		})
		return
	}
	return miss.Richs(path.Join(prefix, kit.Keys(chain)), cache, raw, cb)

}
func Rich(m *ice.Message, prefix string, chain Any, data Any) string {
	cache := Confm(m, prefix, chain)
	kit.If(cache == nil, func() { cache = kit.Data(); m.Confv(prefix, chain, cache) })
	return miss.Rich(path.Join(prefix, kit.Keys(chain)), cache, data)
}
func sortByField(m *ice.Message, fields string, arg ...string) *ice.Message {
	return m.Table(func(value ice.Maps) {
		m.SetAppend().OptionFields(ice.FIELDS_DETAIL)
		kit.For(kit.Split(fields), func(key string) {
			key = strings.TrimSuffix(key, "*")
			if key == HASH {
				m.Push(key, kit.Select(value[key], arg, 0))
			} else {
				m.Push(key, value[key])
			}
			delete(value, key)
		})
		kit.For(kit.SortedKey(value), func(k string) { m.Push(k, value[k]) })
	})
}
