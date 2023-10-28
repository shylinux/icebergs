package mdb

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

type Any = ice.Any
type List = ice.List
type Maps = ice.Maps
type Map = ice.Map

func _mdb_modify(m *ice.Message, value Map, field string, arg ...string) {
	value = kit.GetMeta(value)
	kit.For(arg, func(k, v string) { kit.If(k != field, func() { kit.Value(value, k, v) }) })
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
		cb(ToMaps(value))
	case string, []string, []Any, nil:
		if m.FieldsIsDetail() {
			m.Push(ice.FIELDS_DETAIL, value)
		} else {
			m.Push(key, value, fields, val)
		}
	default:
		m.ErrorNotImplement(cb)
	}
}
func _mdb_export_file(m *ice.Message, arg ...string) string {
	if len(arg) > 3 && strings.Contains(arg[3], ice.PS) {
		return arg[3]
	}
	return path.Join(ice.USR_LOCAL_EXPORT, path.Join(arg[:2]...), arg[2])
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

	FILE = kit.MDB_FILE
	LINK = kit.MDB_LINK
	SCAN = kit.MDB_SCAN
	SHOW = kit.MDB_SHOW
	HELP = kit.MDB_HELP
	ICON = kit.MDB_ICON

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

	TOOLS     = "tools"
	SOURCE    = "_source"
	TARGET    = "_target"
	IMPORTANT = "important"
)
const (
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

	DETAIL = "detail"
	FIELDS = "fields"
	PARAMS = "params"
	OFFEND = "offend"
	OFFSET = "offset"
	RANDOM = "random"
	SUBKEY = "mdb.sub"

	ACTION = "action"
	UPLOAD = "upload"
	RECENT = "recent"
	REPEAT = "repeat"
	REVERT = "revert"
	RENAME = "rename"
	ORDER  = "order"
	PRUNE  = "prune"

	PAGE = "page"
	NEXT = "next"
	PREV = "prev"
	PLAY = "play"

	SORT = "sort"
	JSON = "json"
	CSV  = "csv"
	SUB  = "sub"

	QS = ice.QS
	EQ = ice.EQ
	AT = ice.AT
	FS = ice.FS
)

const MDB = "mdb"

var Index = &ice.Context{Name: MDB, Help: "数据模块", Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {}},
	INPUTS: {Name: "inputs key sub type field value", Hand: func(m *ice.Message, arg ...string) {
		kit.Switch(arg[2],
			HASH, func() { _hash_inputs(m, arg[0], arg[1], kit.Select(NAME, arg, 3), kit.Select("", arg, 4)) },
			ZONE, func() { _zone_inputs(m, arg[0], arg[1], arg[3], kit.Select(NAME, arg, 4), kit.Select("", arg, 5)) },
			LIST, func() { _list_inputs(m, arg[0], arg[1], kit.Select(NAME, arg, 3), kit.Select("", arg, 4)) },
		)
		for _, inputs := range ice.Info.Inputs {
			if arg[2] == ZONE {
				inputs(m, arg[4])
			} else {
				inputs(m, arg[3])
			}
		}
	}},
	INSERT: {Name: "insert key sub type arg...", Hand: func(m *ice.Message, arg ...string) {
		kit.Switch(arg[2],
			HASH, func() { _hash_insert(m, arg[0], arg[1], arg[3:]...) },
			ZONE, func() {
				if arg[3] == ZONE {
					_zone_insert(m, arg[0], arg[1], arg[4], arg[5:]...)
				} else {
					_zone_insert(m, arg[0], arg[1], arg[3], arg[4:]...)
				}
			},
			LIST, func() { _list_insert(m, arg[0], arg[1], arg[3:]...) },
		)
	}},
	DELETE: {Name: "delete key sub type field value", Hand: func(m *ice.Message, arg ...string) {
		kit.Switch(arg[2],
			HASH, func() { _hash_delete(m, arg[0], arg[1], arg[3], arg[4]) },
			// ZONE, func() { _list_delete(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4], arg[5]) },
			// LIST, func() { _list_delete(m, arg[0], arg[1], arg[3], arg[4]) },
		)
	}},
	MODIFY: {Name: "modify key sub type field value arg...", Hand: func(m *ice.Message, arg ...string) {
		kit.Switch(arg[2],
			HASH, func() { _hash_modify(m, arg[0], arg[1], arg[3], arg[4], arg[5:]...) },
			ZONE, func() { _zone_modify(m, arg[0], arg[1], arg[3], arg[4], arg[5:]...) },
			LIST, func() { _list_modify(m, arg[0], arg[1], arg[3], arg[4], arg[5:]...) },
		)
	}},
	SELECT: {Name: "select key sub type field value", Hand: func(m *ice.Message, arg ...string) {
		kit.Switch(arg[2],
			HASH, func() { _hash_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select(FOREACH, arg, 4)) },
			ZONE, func() { _zone_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select("", arg, 4)) },
			LIST, func() { _list_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select("", arg, 4)) },
		)
	}},
	PRUNES: {Name: "prunes key sub type [field value]...", Hand: func(m *ice.Message, arg ...string) {
		kit.Switch(arg[2],
			HASH, func() {
				_hash_prunes(m, arg[0], arg[1], arg[3:]...)
				m.Table(func(value Maps) { _hash_delete(m, arg[0], arg[1], HASH, value[HASH]) })
			},
			// ZONE, func() { _list_prunes(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4:]...) },
			// LIST, func() { _list_prunes(m, arg[0], arg[1], arg[3:]...) },
		)
	}},
	EXPORT: {Name: "export key sub type file", Hand: func(m *ice.Message, arg ...string) {
		m.OptionDefault(CACHE_LIMIT, "-1")
		file := _mdb_export_file(m, arg...)
		kit.Switch(arg[2],
			HASH, func() { _hash_export(m, arg[0], arg[1], file) },
			ZONE, func() { _zone_export(m, arg[0], arg[1], file) },
			LIST, func() { _list_export(m, arg[0], arg[1], file) },
		)
	}},
	IMPORT: {Name: "import key sub type file", Hand: func(m *ice.Message, arg ...string) {
		file := _mdb_export_file(m, arg...)
		kit.Switch(arg[2],
			HASH, func() { _hash_import(m, arg[0], arg[1], file) },
			ZONE, func() { _zone_import(m, arg[0], arg[1], file) },
			LIST, func() { _list_import(m, arg[0], arg[1], file) },
		)
	}},
}}

func init() {
	ice.Index.Register(Index, nil, INPUTS, INSERT, DELETE, MODIFY, SELECT, PRUNES, EXPORT, IMPORT, PLUGIN, RENDER, ENGINE, SEARCH)
}
func init() {
	ice.Module(MDB,
		HashInputs, HashCreate, HashRemove, func(m *ice.Message) { HashPrunes(m, nil) }, HashModify, HashSelect,
		ZoneInputs, ZoneCreate, ZoneRemove, ZoneInsert, ZoneModify, ZoneSelect,
	)
}

func AutoConfig(arg ...Any) *ice.Action {
	return &ice.Action{Hand: func(m *ice.Message, args ...string) {
		if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
			cs[m.CommandKey()] = &ice.Config{Value: kit.Data(arg...)}
		} else {
			kit.For(kit.Dict(arg...), func(k string, v Any) { Config(m, k, v) })
		}
		if cmd := m.Target().Commands[m.CommandKey()]; cmd == nil {
			return
		} else {
			s := Config(m, SHORT)
			kit.If(s == "" || s == UNIQ || strings.Contains(s, ","), func() { s = HASH })
			if cmd.Name == "" {
				cmd.Name = kit.Format("%s %s auto", m.CommandKey(), s)
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
			if cmd.Actions[INSERT] != nil {
				kit.If(cmd.Meta[INSERT] == nil, func() { m.Design(INSERT, "", add(kit.Simple(Config(m, SHORT), kit.Split(ListField(m))))...) })
				kit.If(cmd.Meta[CREATE] == nil, func() { m.Design(CREATE, "", add(kit.Split(Config(m, SHORT)))...) })
			} else if cmd.Actions[CREATE] != nil {
				kit.If(cmd.Meta[CREATE] == nil, func() { m.Design(CREATE, "", add(kit.Split(HashField(m)))...) })
			}
		}
	}}
}
func ImportantZoneAction(arg ...Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { Config(m, IMPORTANT, ice.TRUE) }},
	}, ZoneAction(arg...))
}
func ImportantHashAction(arg ...Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { Config(m, IMPORTANT, ice.TRUE) }},
	}, HashAction(arg...))
}
func saveImportant(m *ice.Message, key, sub string, arg ...string) {
	if m.Option("skip.important") == ice.TRUE {
		return
	}
	kit.If(m.Conf(key, kit.Keys(META, IMPORTANT)) == ice.TRUE, func() { ice.SaveImportant(m, arg...) })
}
func ToMaps(value Map) Maps {
	res := Maps{}
	kit.For(value, func(k, v string) { res[k] = v })
	return res
}
