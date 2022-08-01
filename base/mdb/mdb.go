package mdb

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

type Any = interface{}
type Map = map[string]Any
type Maps = map[string]string

func _domain_chain(m *ice.Message, chain string) string {
	return kit.Keys(m.Option(ice.MSG_DOMAIN), chain)
}
func _file_name(m *ice.Message, arg ...string) string {
	if len(arg) > 3 && strings.Contains(arg[3], ice.PS) {
		return arg[3]
	}
	return path.Join(ice.USR_LOCAL_EXPORT, m.Option(ice.MSG_DOMAIN), path.Join(arg[:2]...), arg[2])
}
func _mdb_args(m *ice.Message, field string, arg ...Any) []string {
	args := kit.Simple(arg...)
	for i := 0; i < len(args); i += 2 {
		if !strings.Contains(field, args[i]) && !strings.HasPrefix(args[i], EXTRA) {
			args[i] = kit.Keys(EXTRA, args[i])
		}
	}
	return args
}
func _mdb_modify(m *ice.Message, val ice.Map, field string, arg ...string) {
	val = kit.GetMeta(val)
	for i := 0; i < len(arg); i += 2 {
		if arg[i] == field {
			continue
		}
		kit.Value(val, arg[i], kit.Select("", arg, i+1))
	}
}
func _mdb_select(m *ice.Message, key string, value ice.Map, fields []string, val ice.Map) {
	switch value = kit.GetMeta(value); cb := m.OptionCB(SELECT).(type) {
	case func(string, []string, ice.Map, ice.Map):
		cb(key, fields, value, val)
	case func([]string, ice.Map):
		cb(fields, value)
	case func(string, ice.Map, ice.Map):
		cb(key, value, val)
	case func(string, ice.Map):
		cb(key, value)
	case func(ice.Map):
		cb(value)
	case func(Any):
		cb(value[TARGET])
	case func(ice.Maps):
		res := ice.Maps{}
		for k, v := range value {
			res[k] = kit.Format(v)
		}
		cb(res)
	case nil:
		if m.FieldsIsDetail() {
			m.Push(ice.CACHE_DETAIL, value)
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
	// 数据
	ID   = kit.MDB_ID
	KEY  = kit.MDB_KEY
	TIME = kit.MDB_TIME
	TYPE = kit.MDB_TYPE
	NAME = kit.MDB_NAME
	TEXT = kit.MDB_TEXT

	// 文件
	LINK = kit.MDB_LINK
	SCAN = kit.MDB_SCAN
	SHOW = kit.MDB_SHOW
	HELP = kit.MDB_HELP
	DATA = kit.MDB_DATA
	FILE = kit.MDB_FILE

	// 存储
	SHORT = kit.MDB_SHORT
	FIELD = kit.MDB_FIELD
	TOTAL = kit.MDB_TOTAL
	COUNT = kit.MDB_COUNT
	LIMIT = kit.MDB_LIMIT
	LEAST = kit.MDB_LEAST
	STORE = kit.MDB_STORE
	FSIZE = kit.MDB_FSIZE

	// 索引
	INDEX  = kit.MDB_INDEX
	VALUE  = kit.MDB_VALUE
	EXTRA  = kit.MDB_EXTRA
	ALIAS  = kit.MDB_ALIAS
	EXPIRE = kit.MDB_EXPIRE
	STATUS = kit.MDB_STATUS
	STREAM = kit.MDB_STREAM
)
const (
	DETAIL = "detail"
	RANDOM = "random"
	ACTION = "action"

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
const (
	CACHE_CLEAR_ON_EXIT = "cache.clear.on.exit"

	SOURCE = "_source"
	TARGET = "_target"
)

const MDB = "mdb"

var Index = &ice.Context{Name: MDB, Help: "数据模块", Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {}},
	INPUTS: {Name: "inputs key sub type field value", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
		switch arg[3] = strings.TrimPrefix(arg[3], "extra."); arg[3] {
		case ice.POD:
			m.Cmdy("route")
		case ice.CTX:
			m.Cmdy("context")
		case ice.CMD:
			m.Cmdy("context", kit.Select(m.Option(ice.CTX), m.Option(kit.Keys(EXTRA, ice.CTX))), "command")
		case "index":
			m.OptionFields(arg[0])
			m.Cmdy("command", SEARCH, "command", kit.Select("", arg, 1))
		default:
			switch arg[2] {
			case ZONE: // inputs key sub type zone field value
				_zone_inputs(m, arg[0], _domain_chain(m, arg[1]), arg[3], kit.Select(NAME, arg, 4), kit.Select("", arg, 5))
			case HASH:
				_hash_inputs(m, arg[0], _domain_chain(m, arg[1]), kit.Select(NAME, arg, 3), kit.Select("", arg, 4))
			case LIST:
				_list_inputs(m, arg[0], _domain_chain(m, arg[1]), kit.Select(NAME, arg, 3), kit.Select("", arg, 4))
			}
		}
	}},
	INSERT: {Name: "insert key sub type arg...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
		defer m.ProcessRefresh3ms()
		switch arg[2] {
		case ZONE: // insert key sub type zone arg...
			_zone_insert(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4:]...)
		case HASH:
			_hash_insert(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		case LIST:
			_list_insert(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		}
	}},
	DELETE: {Name: "delete key sub type field value", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
		defer m.ProcessRefresh3ms()
		switch arg[2] {
		case ZONE: // delete key sub type zone field value
			// _list_delete(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4], arg[5])
		case HASH:
			_hash_delete(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4])
		case LIST:
			// _list_delete(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4])
		}
	}},
	MODIFY: {Name: "modify key sub type field value arg...", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE: // modify key sub type zone id field value
			_zone_modify(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4], arg[5:]...)
		case HASH:
			_hash_modify(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4], arg[5:]...)
		case LIST:
			_list_modify(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4], arg[5:]...)
		}
	}},
	SELECT: {Name: "select key sub type field value", Help: "查询", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE:
			_zone_select(m, arg[0], _domain_chain(m, arg[1]), kit.Select("", arg, 3), kit.Select("", arg, 4))
		case HASH:
			_hash_select(m, arg[0], _domain_chain(m, arg[1]), kit.Select("", arg, 3), kit.Select(FOREACH, arg, 4))
		case LIST:
			_list_select(m, arg[0], _domain_chain(m, arg[1]), kit.Select("", arg, 3), kit.Select("", arg, 4))
		}
	}},
	PRUNES: {Name: "prunes key sub type [field value]...", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE: // prunes key sub type zone field value
			// _list_prunes(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4:]...)
		case HASH:
			_hash_prunes(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
			m.Tables(func(value ice.Maps) { _hash_delete(m, arg[0], _domain_chain(m, arg[1]), HASH, value[HASH]) })
		case LIST:
			// _list_prunes(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		}
	}},
	EXPORT: {Name: "export key sub type file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
		if m.Option(ice.CACHE_LIMIT) == "" {
			m.Option(ice.CACHE_LIMIT, "-1")
		}
		switch file := _file_name(m, arg...); arg[2] {
		case ZONE:
			m.OptionFields(ZoneShort(m), m.Config(FIELD))
			_zone_export(m, arg[0], _domain_chain(m, arg[1]), file)
		case HASH:
			_hash_export(m, arg[0], _domain_chain(m, arg[1]), file)
		case LIST:
			m.OptionFields(m.Config(FIELD))
			_list_export(m, arg[0], _domain_chain(m, arg[1]), file)
		}
	}},
	IMPORT: {Name: "import key sub type file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
		switch file := _file_name(m, arg...); arg[2] {
		case ZONE:
			_zone_import(m, arg[0], _domain_chain(m, arg[1]), file)
		case HASH:
			_hash_import(m, arg[0], _domain_chain(m, arg[1]), file)
		case LIST:
			_list_import(m, arg[0], _domain_chain(m, arg[1]), file)
		}
	}},
}}

func init() {
	ice.Index.Register(Index, nil,
		INSERT, DELETE, MODIFY, SELECT,
		INPUTS, PRUNES, EXPORT, IMPORT,
		SEARCH, ENGINE, PLUGIN, RENDER,
	)
}
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
