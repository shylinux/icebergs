package mdb

import (
	"path"
	"strings"
	"sync"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/task"
)

type Any = interface{}
type Map = map[string]Any
type Maps = map[string]string
type List = []interface{}

func _file_name(m *ice.Message, arg ...string) string {
	if len(arg) > 3 && strings.Contains(arg[3], ice.PS) {
		return arg[3]
	}
	return path.Join(ice.USR_LOCAL_EXPORT, path.Join(arg[:2]...), arg[2])
}
func _mdb_getmeta(m *ice.Message, prefix, chain, key string) string {
	defer RLock(m, prefix, chain)()
	return m.Conf(prefix, kit.Keys(chain, kit.Keym(key)))
}
func _mdb_modify(m *ice.Message, value ice.Map, field string, arg ...string) {
	value = kit.GetMeta(value)
	kit.Fetch(arg, func(k, v string) {
		if k != field {
			kit.Value(value, k, v)
		}
	})
}
func _mdb_select(m *ice.Message, cb Any, key string, value Map, fields []string, val Map) {
	switch value, val = kit.GetMeta(value), kit.GetMeta(val); cb := cb.(type) {
	case func([]string, Map):
		cb(fields, value)
	case func(string, []string, Map, Map):
		cb(key, fields, value, val)
	case func(string, Map, Map):
		cb(key, value, val)
	case func(string, Map):
		cb(key, value)
	case func(Map):
		cb(value)
	case func(Any):
		cb(value[TARGET])
	case func(Maps):
		res := Maps{}
		for k, v := range value {
			res[k] = kit.Format(v)
		}
		cb(res)
	case string, []string, []ice.Any, nil:
		if m.FieldsIsDetail() {
			m.Push(ice.FIELDS_DETAIL, value)
		} else {
			m.Push(key, value, fields, val)
		}
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	DICT = kit.MDB_DICT
	META = kit.MDB_META

	UNIQ    = kit.MDB_UNIQ
	FOREACH = kit.MDB_FOREACH
	RANDOMS = kit.MDB_RANDOMS
)
const (
	ID   = kit.MDB_ID
	KEY  = kit.MDB_KEY
	TIME = kit.MDB_TIME
	TYPE = kit.MDB_TYPE
	NAME = kit.MDB_NAME
	TEXT = kit.MDB_TEXT
	DATA = kit.MDB_DATA
	VIEW = kit.MDB_VIEW

	LINK = kit.MDB_LINK
	FILE = kit.MDB_FILE
	SCAN = kit.MDB_SCAN
	SHOW = kit.MDB_SHOW
	HELP = kit.MDB_HELP

	INDEX  = kit.MDB_INDEX
	VALUE  = kit.MDB_VALUE
	EXTRA  = kit.MDB_EXTRA
	ALIAS  = kit.MDB_ALIAS
	EXPIRE = kit.MDB_EXPIRE
	STATUS = kit.MDB_STATUS
	STREAM = kit.MDB_STREAM

	SHORT = kit.MDB_SHORT
	FIELD = kit.MDB_FIELD
	TOTAL = kit.MDB_TOTAL
	COUNT = kit.MDB_COUNT
	LIMIT = kit.MDB_LIMIT
	LEAST = kit.MDB_LEAST
	STORE = kit.MDB_STORE
	FSIZE = kit.MDB_FSIZE
	TOOLS = "tools"

	SOURCE = "_source"
	TARGET = "_target"

	CACHE_CLEAR_ON_EXIT = "cache.clear.on.exit"
)
const (
	DETAIL = "detail"
	RANDOM = "random"
	ACTION = "action"
	FIELDS = "fields"
	PARAMS = "params"

	RECENT = "recent"
	INPUTS = "inputs"
	CREATE = "create"
	REMOVE = "remove"
	INSERT = "insert"
	DELETE = "delete"
	MODIFY = "modify"
	SELECT = "select"
	PRUNES = "prunes"
	EXPORT = "export"
	IMPORT = "import"

	UPLOAD = "upload"
	REVERT = "revert"
	REPEAT = "repeat"

	NEXT   = "next"
	PREV   = "prev"
	PAGE   = "page"
	OFFEND = "offend"

	JSON = "json"
	CSV  = "csv"
)

const MDB = "mdb"

var Index = &ice.Context{Name: MDB, Help: "数据模块", Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {}},
	INPUTS: {Name: "inputs key sub type field value", Hand: func(m *ice.Message, arg ...string) {
		const (
			SPACE   = "space"
			CONTEXT = "context"
			COMMAND = "command"
			INDEX   = "index"
		)
		switch arg[3] = strings.TrimPrefix(arg[3], EXTRA+ice.PT); arg[3] {
		case ice.POD:
			m.Cmdy(SPACE)
		case ice.CTX:
			m.Cmdy(CONTEXT)
		case ice.CMD:
			m.Cmdy(CONTEXT, kit.Select(m.Option(ice.CTX), m.Option(kit.Keys(EXTRA, ice.CTX))), COMMAND)
		case INDEX:
			m.Cmdy(COMMAND, SEARCH, COMMAND, kit.Select("", arg, 1), ice.OptionFields(arg[3]))
		default:
			switch arg[2] {
			case ZONE:
				_zone_inputs(m, arg[0], arg[1], arg[3], kit.Select(NAME, arg, 4), kit.Select("", arg, 5))
			case HASH:
				_hash_inputs(m, arg[0], arg[1], kit.Select(NAME, arg, 3), kit.Select("", arg, 4))
			case LIST:
				_list_inputs(m, arg[0], arg[1], kit.Select(NAME, arg, 3), kit.Select("", arg, 4))
			}
		}
	}},
	INSERT: {Name: "insert key sub type arg...", Hand: func(m *ice.Message, arg ...string) {
		defer m.ProcessRefresh()
		switch arg[2] {
		case ZONE:
			_zone_insert(m, arg[0], arg[1], arg[3], arg[4:]...)
		case HASH:
			_hash_insert(m, arg[0], arg[1], arg[3:]...)
		case LIST:
			_list_insert(m, arg[0], arg[1], arg[3:]...)
		}
	}},
	DELETE: {Name: "delete key sub type field value", Hand: func(m *ice.Message, arg ...string) {
		defer m.ProcessRefresh()
		switch arg[2] {
		case ZONE:
			// _list_delete(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4], arg[5])
		case HASH:
			_hash_delete(m, arg[0], arg[1], arg[3], arg[4])
		case LIST:
			// _list_delete(m, arg[0], arg[1], arg[3], arg[4])
		}
	}},
	MODIFY: {Name: "modify key sub type field value arg...", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE:
			_zone_modify(m, arg[0], arg[1], arg[3], arg[4], arg[5:]...)
		case HASH:
			_hash_modify(m, arg[0], arg[1], arg[3], arg[4], arg[5:]...)
		case LIST:
			_list_modify(m, arg[0], arg[1], arg[3], arg[4], arg[5:]...)
		}
	}},
	SELECT: {Name: "select key sub type field value", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE:
			_zone_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select("", arg, 4))
		case HASH:
			_hash_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select(FOREACH, arg, 4))
		case LIST:
			_list_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select("", arg, 4))
		}
	}},
	PRUNES: {Name: "prunes key sub type [field value]...", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE:
			// _list_prunes(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4:]...)
		case HASH:
			_hash_prunes(m, arg[0], arg[1], arg[3:]...)
			m.Tables(func(value ice.Maps) { _hash_delete(m, arg[0], arg[1], HASH, value[HASH]) })
		case LIST:
			// _list_prunes(m, arg[0], arg[1], arg[3:]...)
		}
	}},
	EXPORT: {Name: "export key sub type file", Hand: func(m *ice.Message, arg ...string) {
		m.OptionDefault(CACHE_LIMIT, "-1")
		switch file := _file_name(m, arg...); arg[2] {
		case ZONE:
			_zone_export(m, arg[0], arg[1], file)
		case HASH:
			_hash_export(m, arg[0], arg[1], file)
		case LIST:
			_list_export(m, arg[0], arg[1], file)
		}
		m.StatusTime(LINK, "/share/local/"+m.Result()).Process("_clear")
	}},
	IMPORT: {Name: "import key sub type file", Hand: func(m *ice.Message, arg ...string) {
		switch file := _file_name(m, arg...); arg[2] {
		case ZONE:
			_zone_import(m, arg[0], arg[1], file)
		case HASH:
			_hash_import(m, arg[0], arg[1], file)
		case LIST:
			_list_import(m, arg[0], arg[1], file)
		}
	}},
}}

func init() {
	ice.Index.Register(Index, nil, INSERT, DELETE, MODIFY, SELECT, INPUTS, PRUNES, EXPORT, IMPORT, PLUGIN, RENDER, ENGINE, SEARCH)
}
func AutoConfig(args ...ice.Any) *ice.Action {
	return &ice.Action{Hand: func(m *ice.Message, arg ...string) {
		if cs := m.Target().Configs; len(args) > 0 {
			if cs[m.CommandKey()] == nil {
				cs[m.CommandKey()] = &ice.Config{Value: kit.Data(args...)}
				// ice.Info.Load(m, m.CommandKey())
			} else {
				for k, v := range kit.Dict(args...) {
					m.Config(k, v)
				}
			}
		}
		if cmd := m.Target().Commands[m.CommandKey()]; cmd == nil {
			return
		} else if cmd.Actions[INSERT] != nil {
			if inputs := []ice.Any{}; cmd.Meta[INSERT] == nil {
				kit.Fetch(kit.Filters(kit.Simple(m.Config(SHORT), kit.Split(ListField(m))), "", TIME, ID), func(k string) { inputs = append(inputs, k) })
				m.Design(INSERT, "添加", inputs...)
			}
			if inputs := []ice.Any{}; cmd.Meta[CREATE] == nil {
				kit.Fetch(kit.Filters(kit.Split(kit.Select(m.Config(SHORT), m.Config(FIELDS))), TIME, HASH, COUNT), func(k string) { inputs = append(inputs, k) })
				m.Design(CREATE, "创建", inputs...)
			}
		} else if cmd.Actions[CREATE] != nil {
			if inputs := []ice.Any{}; cmd.Meta[CREATE] == nil {
				kit.Fetch(kit.Filters(kit.Split(HashField(m)), TIME, HASH), func(k string) { inputs = append(inputs, k) })
				m.Design(CREATE, "创建", inputs...)
			}
		}
	}}
}

var _lock = task.Lock{}
var _locks = map[string]*task.Lock{}

func getLock(m *ice.Message, key string) *task.Lock {
	if key == "" {
		key = m.PrefixKey()
	}
	defer _lock.Lock()()
	l, ok := _locks[key]
	if !ok {
		l = &task.Lock{}
		_locks[key] = l
	}
	return l
}
func Lock(m *ice.Message, arg ...ice.Any) func()  { return getLock(m, kit.Keys(arg...)).Lock() }
func RLock(m *ice.Message, arg ...ice.Any) func() { return getLock(m, kit.Keys(arg...)).RLock() }

func Config(m *ice.Message, key string, arg ...ice.Any) string {
	if len(arg) > 0 {
		defer Lock(m, m.PrefixKey(), key)()
	} else {
		defer RLock(m, m.PrefixKey(), key)()
	}
	return m.Config(key, arg...)
}

var cache = sync.Map{}

func Cache(m *ice.Message, key string, add func() ice.Any) ice.Any {
	if add == nil {
		cache.Delete(key)
		return nil
	}
	if val, ok := cache.Load(key); ok {
		return val
	}
	if val := add(); val != nil {
		cache.Store(key, val)
		return val
	}
	return nil
}
