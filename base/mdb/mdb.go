package mdb

import (
	"encoding/csv"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _file_name(m *ice.Message, arg ...string) string {
	return kit.Select(path.Join(m.Option(ice.MSG_LOCAL), ice.USR_LOCAL, EXPORT, path.Join(arg[:2]...), arg[2]), arg, 3)
}
func _domain_chain(m *ice.Message, chain string) string {
	return kit.Keys(m.Option(ice.MSG_DOMAIN), chain)
}

func _zone_fields(m *ice.Message) []string {
	return kit.Split(kit.Select("zone,id,time,type,name,text", strings.Join(kit.Simple(m.Optionv(FIELDS)), ",")))
}
func _zone_select(m *ice.Message, prefix, chain, zone string, id string) {
	if zone == RANDOM {
		zone = kit.MDB_RANDOMS
	}

	fields := _zone_fields(m)
	cb := m.Optionv(kit.Keycb(SELECT))
	m.Richs(prefix, chain, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		val = kit.GetMeta(val)
		if zone == "" {
			if m.Option(FIELDS) == DETAIL {
				m.Push(DETAIL, val)
			} else {
				m.Push(key, val, fields)
			}
			return
		}

		m.Grows(prefix, kit.Keys(chain, kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			value = kit.GetMeta(value)

			switch cb := cb.(type) {
			case func(string, []string, map[string]interface{}, map[string]interface{}):
				cb(key, fields, value, val)
			case func(string, map[string]interface{}, map[string]interface{}):
				cb(key, value, val)
			case func(string, map[string]interface{}):
				cb(key, value)
			default:
				if m.Option(FIELDS) == DETAIL {
					m.Push(DETAIL, value)
				} else {
					m.Push(key, value, fields, val)
				}
			}
		})
	})
}
func _zone_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := kit.Create(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	fields := _zone_fields(m)
	w.Write(fields)

	count := 0
	m.Richs(prefix, chain, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		val = kit.GetMeta(val)

		m.Grows(prefix, kit.Keys(chain, kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			value = kit.GetMeta(value)

			list := []string{}
			for _, k := range fields {
				list = append(list, kit.Select(kit.Format(kit.Value(val, k)), kit.Format(kit.Value(value, k))))
			}
			w.Write(list)
			count++
		})
	})

	m.Log_EXPORT(kit.MDB_KEY, path.Join(prefix, chain), kit.MDB_FILE, p, kit.MDB_COUNT, count)
	m.Conf(prefix, kit.Keys(chain, kit.MDB_HASH), "")
	m.Echo(p)
}
func _zone_import(m *ice.Message, prefix, chain, file string) {
	f, e := os.Open(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	r := csv.NewReader(f)
	head, _ := r.Read()
	count := 0

	list := map[string]string{}
	zkey := kit.Select(head[0], m.Option(FIELDS))

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
			case kit.MDB_ID:
				continue
			case kit.MDB_EXTRA:
				kit.Value(data, k, kit.UnMarshal(line[i]))
			default:
				kit.Value(data, k, line[i])
			}
		}
		if list[zone] == "" {
			list[zone] = m.Rich(prefix, chain, kit.Data(zkey, zone))
		}

		m.Grow(prefix, kit.Keys(chain, kit.MDB_HASH, list[zone]), data)
		count++
	}

	m.Log_IMPORT(kit.MDB_KEY, path.Join(prefix, chain), kit.MDB_COUNT, count)
	m.Echo("%d", count)
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
