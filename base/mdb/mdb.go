package mdb

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"encoding/csv"
	"encoding/json"
	"os"
	"path"
	"sort"
)

func _file_name(m *ice.Message, arg ...string) string {
	return kit.Select(path.Join("usr/export", kit.Select(arg[0], arg[0]+"/"+arg[1], arg[1] != ""), arg[2]), arg, 3)
}

func _hash_insert(m *ice.Message, prefix, chain string, arg ...string) {
	m.Log_INSERT("prefix", prefix, arg[0], arg[1])
	m.Echo(m.Rich(prefix, chain, kit.Data(arg)))

}
func _hash_delete(m *ice.Message, prefix, chain, field, value string) {
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
		m.Log_DELETE("prefix", prefix, field, value, "value", kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, kit.MDB_HASH, key), "")
	})
}
func _hash_select(m *ice.Message, prefix, chain, field, value string) {
	if field == kit.MDB_HASH && value == "random" {
		value = kit.MDB_RANDOMS
	}
	fields := kit.Split(kit.Select("time,hash,type,name,text", m.Option(FIELDS)))
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}
		if m.Option(FIELDS) == DETAIL {
			m.Push(DETAIL, val)
		} else {
			m.Push(key, val, fields)
		}
	})
	if m.Option(FIELDS) != DETAIL {
		m.Sort(kit.MDB_TIME, "time_r")
	}
}
func _hash_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Richs(prefix, chain, value, func(key string, value map[string]interface{}) {
		if value[kit.MDB_META] != nil {
			value = value[kit.MDB_META].(map[string]interface{})
		}
		for i := 0; i < len(arg)-1; i += 2 {
			if arg[i] == field {
				continue
			}
			kit.Value(value, arg[i], arg[i+1])
		}
		m.Log_MODIFY("prefix", prefix, field, value, arg)
	})
}
func _hash_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := kit.Create(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	en.Encode(m.Confv(prefix, kit.Keys(chain, HASH)))

	m.Log_EXPORT(kit.MDB_FILE, p)
	m.Echo(p)
}
func _hash_import(m *ice.Message, prefix, chain, file string) {
	f, e := os.Open(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	list := map[string]interface{}{}
	de := json.NewDecoder(f)
	de.Decode(&list)

	count := 0
	if m.Conf(prefix, kit.Keys(chain, kit.MDB_META, kit.MDB_SHORT)) == "" {
		for k, data := range list {
			// 导入数据
			m.Conf(prefix, kit.Keys(chain, kit.MDB_HASH, k), data)
			count++
		}
	} else {
		for _, data := range list {
			// 导入数据
			m.Rich(prefix, chain, data)
			count++
		}
	}

	m.Log_IMPORT(kit.MDB_KEY, kit.Keys(prefix, chain), kit.MDB_COUNT, count)
	m.Echo("%d", count)
}
func _hash_prunes(m *ice.Message, prefix, chain string, arg ...string) {
	m.Richs(prefix, chain, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}
		for i := 0; i < len(arg)-1; i += 2 {
			if val[arg[i]] != arg[i+1] {
				return
			}
		}
		_hash_delete(m, prefix, chain, kit.MDB_HASH, key)
		m.Push(key, val)
	})
}
func _hash_inputs(m *ice.Message, prefix, chain string, field, value string) {
	list := map[string]int{}
	m.Richs(prefix, chain, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}
		if field == kit.MDB_HASH {
			list[key]++
		} else {
			list[kit.Format(val[field])]++
		}
	})
	for k, i := range list {
		m.Push(field, k)
		m.Push(kit.MDB_COUNT, i)
	}
	m.Sort(kit.MDB_COUNT, "int_r")
}

func _list_insert(m *ice.Message, prefix, chain string, arg ...string) {
	m.Log_INSERT("prefix", prefix, arg[0], arg[1])
	m.Echo("%d", m.Grow(prefix, chain, kit.Dict(arg)))
}
func _list_delete(m *ice.Message, prefix, chain, field, value string) {
}
func _list_select(m *ice.Message, prefix, chain, field, value string) {
	fields := kit.Split(kit.Select("time,id,type,name,text", m.Option(FIELDS)), ",")
	m.Grows(prefix, chain, field, value, func(index int, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}

		if m.Option(FIELDS) == DETAIL {
			m.Push(DETAIL, val)
		} else {
			m.Push(kit.Format(index), val, fields)
		}
	})
	if m.Option(FIELDS) != DETAIL {
		m.Sort(kit.MDB_ID, "int_r")
	}
}
func _list_modify(m *ice.Message, prefix, chain string, field, value string, arg ...string) {
	m.Grows(prefix, chain, field, value, func(index int, value map[string]interface{}) {
		if value[kit.MDB_META] != nil {
			value = value[kit.MDB_META].(map[string]interface{})
		}
		for i := 0; i < len(arg)-1; i += 2 {
			if arg[i] == field {
				continue
			}
			kit.Value(value, arg[i], arg[i+1])
		}
		m.Log_MODIFY("prefix", prefix, field, value, arg)
	})
}
func _list_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := kit.Create(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	count := 0
	head := []string{}
	m.Grows(prefix, chain, "", "", func(index int, value map[string]interface{}) {
		if index == 0 {
			// 输出表头
			for k := range value {
				head = append(head, k)
			}
			sort.Strings(head)
			w.Write(head)
		}

		// 输出数据
		data := []string{}
		for _, k := range head {
			data = append(data, kit.Format(value[k]))
		}
		w.Write(data)
		count++
	})

	m.Log_EXPORT(kit.MDB_FILE, p, kit.MDB_COUNT, count)
	m.Echo(p)
}
func _list_import(m *ice.Message, prefix, chain, file string) {
	f, e := os.Open(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	r := csv.NewReader(f)

	count := 0
	head, _ := r.Read()
	for {
		line, e := r.Read()
		if e != nil {
			break
		}

		data := kit.Dict()
		for i, k := range head {
			if k == kit.MDB_EXTRA {
				data[k] = kit.UnMarshal(line[i])
			} else {
				data[k] = line[i]
			}
		}

		// 导入数据
		m.Grow(prefix, chain, data)
		count++
	}

	m.Log_IMPORT(kit.MDB_KEY, kit.Keys(prefix, chain), kit.MDB_COUNT, count)
	m.Echo("%d", count)
}
func _list_inputs(m *ice.Message, prefix, chain string, field, value string) {
	list := map[string]int{}
	m.Grows(prefix, chain, "", "", func(index int, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}
		list[kit.Format(val[field])]++
	})
	for k, i := range list {
		m.Push(field, k)
		m.Push(kit.MDB_COUNT, i)
	}
	m.Sort(kit.MDB_COUNT, "int_r")
}

func _zone_select(m *ice.Message, prefix, chain, zone string, id string) {
	cb := m.Optionv("cache.cb")
	fields := kit.Split(kit.Select("zone,id,time,type,name,text", m.Option(FIELDS)))
	m.Richs(prefix, chain, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}
		if zone == "" {
			m.Push(key, val, fields)
			return
		}

		m.Grows(prefix, kit.Keys(chain, kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			if value[kit.MDB_META] != nil {
				value = value[kit.MDB_META].(map[string]interface{})
			}

			switch cb := cb.(type) {
			case func(string, map[string]interface{}, map[string]interface{}):
				cb(key, value, val)
			default:
				if len(fields) == 1 && fields[0] == DETAIL {
					m.Push(DETAIL, value)
					break
				}
				m.Push(key, value, fields, val)
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

	fields := kit.Split(kit.Select("zone,id,time,type,name,text", m.Option(FIELDS)))
	m.Assert(w.Write(fields))

	count := 0
	m.Richs(prefix, chain, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}

		m.Grows(prefix, kit.Keys(chain, kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			if value[kit.MDB_META] != nil {
				value = value[kit.MDB_META].(map[string]interface{})
			}

			list := []string{}
			for _, k := range fields {
				list = append(list, kit.Select(kit.Format(val[k]), kit.Format(value[k])))
			}
			m.Assert(w.Write(list))
			count++
		})
	})

	m.Log_EXPORT(kit.MDB_FILE, p, kit.MDB_COUNT, count)
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

	zkey := m.Option(FIELDS)

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
			list[zone] = m.Rich(prefix, chain, kit.Dict(zkey, zone))
		}

		m.Grow(prefix, kit.Keys(chain, kit.MDB_HASH, list[zone]), data)
		count++
	}

	m.Log_IMPORT(kit.MDB_KEY, kit.Keys(prefix, chain), kit.MDB_COUNT, count)
	m.Echo("%d", count)
}

const (
	CSV  = "csv"
	JSON = "json"
)
const (
	DICT = "dict"
	META = "meta"
	HASH = "hash"
	LIST = "list"
	ZONE = "zone"
)
const (
	FIELDS = "fields"
	DETAIL = "detail"

	INVITE = "invite"
	COMMIT = "commit"

	CREATE = "create"
	INSERT = "insert"
	MODIFY = "modify"
	SELECT = "select"
	DELETE = "delete"
	REMOVE = "remove"

	EXPORT = "export"
	IMPORT = "import"
	PRUNES = "prunes"
	INPUTS = "inputs"
	SCRIPT = "script"
)

var Index = &ice.Context{Name: "mdb", Help: "数据模块",
	Commands: map[string]*ice.Command{
		INSERT: {Name: "insert conf key type arg...", Help: "添加", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				_hash_insert(m, arg[0], arg[1], arg[3:]...)
			case LIST:
				_list_insert(m, arg[0], arg[1], arg[3:]...)
			}
		}},
		DELETE: {Name: "delete conf key type field value", Help: "删除", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				_hash_delete(m, arg[0], arg[1], arg[3], arg[4])
			case LIST:
				_list_delete(m, arg[0], arg[1], arg[3], arg[4])
			}
		}},
		MODIFY: {Name: "modify conf key type field value arg...", Help: "编辑", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				_hash_modify(m, arg[0], arg[1], arg[3], arg[4], arg[5:]...)
			case LIST:
				_list_modify(m, arg[0], arg[1], arg[3], arg[4], arg[5:]...)
			}
		}},
		SELECT: {Name: "select conf key type field value", Help: "数据查询", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				_hash_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select(kit.MDB_FOREACH, arg, 4))
			case LIST:
				_list_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select("", arg, 4))
			case ZONE:
				_zone_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select("", arg, 4))
			}
		}},
		EXPORT: {Name: "export conf key type file", Help: "导出数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cache.limit", -2)
			switch file := _file_name(m, arg...); arg[2] {
			case HASH:
				_hash_export(m, arg[0], arg[1], file)
			case LIST:
				_list_export(m, arg[0], arg[1], file)
			case ZONE:
				_zone_export(m, arg[0], arg[1], file)
			}
		}},
		IMPORT: {Name: "import conf key type file", Help: "导入数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch file := _file_name(m, arg...); arg[2] {
			case HASH:
				_hash_import(m, arg[0], arg[1], file)
			case LIST:
				_list_import(m, arg[0], arg[1], file)
			case ZONE:
				_zone_import(m, arg[0], arg[1], file)
			}
		}},
		PRUNES: {Name: "prunes conf key type [field value]...", Help: "清理数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				_hash_prunes(m, arg[0], arg[1], arg[3:]...)
			case LIST:
			}
		}},
		INPUTS: {Name: "inputs conf key type field value", Help: "输入补全", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				_hash_inputs(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select("", arg, 4))
			case LIST:
			}
		}},
	},
}

func init() {
	ice.Index.Register(Index, nil,
		INSERT, DELETE, MODIFY, SELECT,
		EXPORT, IMPORT, PRUNES, INPUTS,
		PLUGIN, RENDER, SEARCH, ENGINE,
	)
}
