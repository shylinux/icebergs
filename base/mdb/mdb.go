package mdb

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _file_name(m *ice.Message, arg ...string) string {
	if len(arg) > 3 && strings.Contains(arg[3], ice.PS) {
		return arg[3]
	}
	return path.Join(ice.USR_LOCAL_EXPORT, m.Option(ice.MSG_DOMAIN), path.Join(arg[:2]...), arg[2])
	return kit.Select(path.Join(ice.USR_LOCAL_EXPORT, m.Option(ice.MSG_DOMAIN), path.Join(arg[:2]...), arg[2]), arg, 3)
}
func _domain_chain(m *ice.Message, chain string) string {
	return kit.Keys(m.Option(ice.MSG_DOMAIN), chain)
}

const (
	CSV  = "csv"
	JSON = "json"
)
const (
	DICT = kit.MDB_DICT
	META = kit.MDB_META
	UNIQ = kit.MDB_UNIQ

	FOREACH = kit.MDB_FOREACH
	RANDOMS = kit.MDB_RANDOMS
)
const (
	// 数据
	ID   = kit.MDB_ID
	KEY  = kit.MDB_KEY
	TIME = kit.MDB_TIME
	// ZONE = kit.MDB_ZONE
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

	CREATE = "create"
	REMOVE = "remove"
	INSERT = "insert"
	DELETE = "delete"
	MODIFY = "modify"
	SELECT = "select"

	INPUTS = "inputs"
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
)
const (
	CACHE_CLEAR_ON_EXIT = "cache.clear.on.exit"
)

func PrevPageLimit(m *ice.Message, total string, arg ...string) {
	if kit.Int(kit.Select("0", arg, 1)) > 0 {
		PrevPage(m, total, arg...)
	} else {
		m.Toast("已经是最前一页啦!")
		m.ProcessHold()
	}
}
func PrevPage(m *ice.Message, total string, arg ...string) {
	limit, offend := kit.Select("10", arg, 0), kit.Select("0", arg, 1)
	offends := kit.Int(offend) - kit.Int(limit)
	if total != "0" && (offends <= -kit.Int(total) || offends >= kit.Int(total)) {
		m.Toast("已经是最前一页啦!")
		m.ProcessHold()
		return
	}
	if offends == 0 {
		m.ProcessRewrite("offend", "")
	} else {
		m.ProcessRewrite("offend", offends)
	}

}
func NextPage(m *ice.Message, total string, arg ...string) {
	limit, offend := kit.Select("10", arg, 0), kit.Select("0", arg, 1)
	offends := kit.Int(offend) + kit.Int(limit)
	if total != "0" && (offends <= -kit.Int(total) || offends >= kit.Int(total)) {
		m.Toast("已经是最后一页啦!")
		m.ProcessHold()
		return
	}
	if offends == 0 {
		m.ProcessRewrite("offend", "")
	} else {
		m.ProcessRewrite("offend", offends)
	}
}
func NextPageLimit(m *ice.Message, total string, arg ...string) {
	if kit.Int(kit.Select("0", arg, 1)) < 0 {
		NextPage(m, total, arg...)
	} else {
		m.Toast("已经是最后一页啦!")
		m.ProcessHold()
	}
}

const MDB = "mdb"

var Index = &ice.Context{Name: MDB, Help: "数据模块", Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {}},
	INSERT: {Name: "insert key sub type arg...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE: // insert key sub type zone arg...
			_list_insert(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4:]...)
		case HASH:
			_hash_insert(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		case LIST:
			_list_insert(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		}
		m.ProcessRefresh3ms()
	}},
	DELETE: {Name: "delete key sub type field value", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE: // delete key sub type zone field value
			_list_delete(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4], arg[5])
		case HASH:
			_hash_delete(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4])
		case LIST:
			_list_delete(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4])
		}
		m.ProcessRefresh3ms()
	}},
	MODIFY: {Name: "modify key sub type field value arg...", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE: // modify key sub type zone id field value
			_list_modify(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), ID, arg[4], arg[5:]...)
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
	INPUTS: {Name: "inputs key sub type field value", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
		switch arg[3] {
		case "index":
			m.OptionFields(arg[3])
			m.Cmdy("command", "search", "command")
		}
		switch arg[2] {
		case ZONE: // inputs key sub type zone field value
			_list_inputs(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), kit.Select(NAME, arg, 4), kit.Select("", arg, 5))
		case HASH:
			_hash_inputs(m, arg[0], _domain_chain(m, arg[1]), kit.Select(NAME, arg, 3), kit.Select("", arg, 4))
		case LIST:
			_list_inputs(m, arg[0], _domain_chain(m, arg[1]), kit.Select(NAME, arg, 3), kit.Select("", arg, 4))
		}
	}},
	PRUNES: {Name: "prunes key sub type [field value]...", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
		switch arg[2] {
		case ZONE: // prunes key sub type zone field value
			_list_prunes(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.KeyHash(arg[3]))), arg[4:]...)
		case HASH:
			_hash_prunes(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		case LIST:
			_list_prunes(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		}
	}},
	EXPORT: {Name: "export key sub type file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
		if m.Option(ice.CACHE_LIMIT) == "" {
			m.Option(ice.CACHE_LIMIT, "-1")
		}
		switch file := _file_name(m, arg...); arg[2] {
		case ZONE:
			_zone_export(m, arg[0], _domain_chain(m, arg[1]), file)
		case HASH:
			_hash_export(m, arg[0], _domain_chain(m, arg[1]), file)
		case LIST:
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
