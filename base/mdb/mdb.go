package mdb

import (
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _file_name(m *ice.Message, arg ...string) string {
	return kit.Select(path.Join(m.Option(ice.MSG_LOCAL), ice.USR_LOCAL, EXPORT, path.Join(arg[:2]...), arg[2]), arg, 3)
}
func _domain_chain(m *ice.Message, chain string) string {
	return kit.Keys(m.Option(ice.MSG_DOMAIN), chain)
}

const (
	CSV  = "csv"
	JSON = "json"
)
const (
	DICT = "dict"
	META = "meta"
	ZONE = "zone"
)
const (
	FIELDS = "fields"
	DETAIL = "detail"
	RANDOM = "random"

	CREATE = "create"
	INSERT = "insert"
	MODIFY = "modify"
	SELECT = "select"
	DELETE = "delete"
	REMOVE = "remove"

	EXPORT = "export"
	IMPORT = "import"
	INPUTS = "inputs"
	PRUNES = "prunes"
	REVERT = "revert"
	REPEAT = "repeat"
	UPLOAD = "upload"

	NEXT = "next"
	PREV = "prev"
)
const (
	CACHE_LIMIT  = "cache.limit"
	CACHE_FIELD  = "cache.field"
	CACHE_VALUE  = "cache.value"
	CACHE_OFFEND = "cache.offend"
	CACHE_FILTER = "cache.filter"

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
	if offends <= -kit.Int(total) || offends >= kit.Int(total) {
		m.Toast("已经是最前一页啦!")
		m.ProcessHold()
		return
	}
	m.ProcessRewrite("offend", offends)

}
func NextPage(m *ice.Message, total string, arg ...string) {
	limit, offend := kit.Select("10", arg, 0), kit.Select("0", arg, 1)
	offends := kit.Int(offend) + kit.Int(limit)
	if offends <= -kit.Int(total) || offends >= kit.Int(total) {
		m.Toast("已经是最后一页啦!")
		m.ProcessHold()
		return
	}
	m.ProcessRewrite("offend", offends)
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

var Index = &ice.Context{Name: MDB, Help: "数据模块", Commands: map[string]*ice.Command{
	INSERT: {Name: "insert key sub type arg...", Help: "添加", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		switch arg[2] {
		case ZONE:
			_list_insert(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.SubKey(arg[3]))), arg[4:]...)
		case HASH:
			_hash_insert(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		case LIST:
			_list_insert(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		}
	}},
	DELETE: {Name: "delete key sub type field value", Help: "删除", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		switch arg[2] {
		case ZONE:
			_list_delete(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4])
		case HASH:
			_hash_delete(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4])
		case LIST:
			_list_delete(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4])
		}
	}},
	MODIFY: {Name: "modify key sub type field value arg...", Help: "编辑", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		switch arg[2] {
		case ZONE:
			_list_modify(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.SubKey(arg[3]))), kit.MDB_ID, arg[4], arg[5:]...)
		case HASH:
			_hash_modify(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4], arg[5:]...)
		case LIST:
			_list_modify(m, arg[0], _domain_chain(m, arg[1]), arg[3], arg[4], arg[5:]...)
		}
	}},
	SELECT: {Name: "select key sub type field value", Help: "查询", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		switch arg[2] {
		case ZONE:
			_zone_select(m, arg[0], _domain_chain(m, arg[1]), kit.Select("", arg, 3), kit.Select("", arg, 4))
		case HASH:
			_hash_select(m, arg[0], _domain_chain(m, arg[1]), kit.Select("", arg, 3), kit.Select(kit.MDB_FOREACH, arg, 4))
		case LIST:
			_list_select(m, arg[0], _domain_chain(m, arg[1]), kit.Select("", arg, 3), kit.Select("", arg, 4))
		}
	}},
	EXPORT: {Name: "export key sub type file", Help: "导出", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		switch file := _file_name(m, arg...); arg[2] {
		case ZONE:
			_zone_export(m, arg[0], _domain_chain(m, arg[1]), file)
		case HASH:
			_hash_export(m, arg[0], _domain_chain(m, arg[1]), file)
		case LIST:
			_list_export(m, arg[0], _domain_chain(m, arg[1]), file)
		}
	}},
	IMPORT: {Name: "import key sub type file", Help: "导入", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		switch file := _file_name(m, arg...); arg[2] {
		case ZONE:
			_zone_import(m, arg[0], _domain_chain(m, arg[1]), file)
		case HASH:
			_hash_import(m, arg[0], _domain_chain(m, arg[1]), file)
		case LIST:
			_list_import(m, arg[0], _domain_chain(m, arg[1]), file)
		}
	}},
	INPUTS: {Name: "inputs key sub type field value", Help: "补全", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		switch arg[2] {
		case ZONE:
			_list_inputs(m, arg[0], _domain_chain(m, kit.Keys(arg[1], kit.SubKey(arg[3]))), kit.Select("name", arg, 4), kit.Select("", arg, 5))
		case HASH:
			_hash_inputs(m, arg[0], _domain_chain(m, arg[1]), kit.Select("name", arg, 3), kit.Select("", arg, 4))
		case LIST:
			_list_inputs(m, arg[0], _domain_chain(m, arg[1]), kit.Select("name", arg, 3), kit.Select("", arg, 4))
		}
	}},
	PRUNES: {Name: "prunes key sub type [field value]...", Help: "清理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		switch arg[2] {
		case HASH:
			_hash_prunes(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		case LIST:
			_list_prunes(m, arg[0], _domain_chain(m, arg[1]), arg[3:]...)
		}
	}},
}}

func init() {
	ice.Index.Register(Index, nil,
		INSERT, DELETE, MODIFY, SELECT,
		EXPORT, IMPORT, PRUNES, INPUTS,
		PLUGIN, RENDER, ENGINE, SEARCH,
	)
}
