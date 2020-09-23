package mdb

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/task"

	"encoding/csv"
	"encoding/json"
	"os"
	"path"
	"sort"
	"strings"
)

func _file_name(m *ice.Message, arg ...string) string {
	return kit.Select(path.Join("usr/export", kit.Select(arg[0], arg[0]+"/"+arg[1], arg[1] != ""), arg[2]), arg, 3)
}

func _hash_insert(m *ice.Message, prefix, key string, arg ...string) string {
	m.Log_INSERT("prefix", prefix, arg[0], arg[1])
	return m.Rich(prefix, key, kit.Data(arg))

}
func _hash_delete(m *ice.Message, prefix, chain, field, value string) {
	m.Richs(prefix, chain, value, func(key string, val map[string]interface{}) {
		m.Log_DELETE("prefix", prefix, field, value, "value", kit.Format(val))
		m.Conf(prefix, kit.Keys(chain, kit.MDB_HASH, key), "")
	})
}
func _hash_select(m *ice.Message, prefix, key, field, value string) {
	if field == "hash" && value == "random" {
		value = kit.MDB_RANDOMS
	}
	fields := strings.Split(kit.Select("time,hash,type,name,text", m.Option(FIELDS)), ",")
	m.Richs(prefix, key, value, func(key string, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}
		if field != "" && field != kit.MDB_HASH && value != val[field] && value != kit.MDB_FOREACH {
			return
		}
		if m.Option(FIELDS) == "detail" {
			m.Push("detail", val)
		} else {
			m.Push(key, val, fields, val[kit.MDB_META])
		}
	})
	if m.Option(FIELDS) != "detail" {
		m.Sort(kit.MDB_TIME, "time_r")
	}
}
func _hash_modify(m *ice.Message, prefix, key string, field, value string, arg ...string) {
	m.Richs(prefix, key, value, func(key string, value map[string]interface{}) {
		if value[kit.MDB_META] != nil {
			value = value[kit.MDB_META].(map[string]interface{})
		}
		for i := 0; i < len(arg)-1; i += 2 {
			if arg[i] == field {
				continue
			}
			kit.Value(value, arg[i], arg[i+1])
		}
	})
	m.Log_MODIFY("prefix", prefix, field, value, arg)
}
func _hash_import(m *ice.Message, prefix, key, file string) {
	f, e := os.Open(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	list := map[string]interface{}{}
	de := json.NewDecoder(f)
	de.Decode(&list)

	if m.Conf(prefix, kit.Keys(key, kit.MDB_META, kit.MDB_SHORT)) == "" {
		m.Conf(prefix, kit.Keys(key, kit.MDB_META, kit.MDB_SHORT), m.Conf(prefix, kit.Keys(kit.MDB_META, kit.MDB_SHORT)))
	}

	count := 0
	if m.Conf(prefix, kit.Keys(key, kit.MDB_META, kit.MDB_SHORT)) == "" {
		for k, data := range list {
			// 导入数据
			m.Conf(prefix, kit.Keys(key, kit.MDB_HASH, k), data)
			count++
		}
	} else {
		for _, data := range list {
			// 导入数据
			m.Rich(prefix, key, data)
			count++
		}
	}

	m.Log_IMPORT(kit.MDB_KEY, kit.Keys(prefix, key), kit.MDB_COUNT, count)
	m.Echo(kit.Keys(file, JSON))
}
func _hash_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(kit.Keys(file, JSON))
	m.Assert(e)
	defer f.Close()

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	en.Encode(m.Confv(prefix, kit.Keys(key, HASH)))
	m.Log_EXPORT(kit.MDB_FILE, p)
	m.Echo(p)
}
func _hash_inputs(m *ice.Message, prefix, key string, field, value string) {
	list := map[string]int{}
	m.Richs(prefix, key, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		list[kit.Format(val[field])]++
	})
	for k, i := range list {
		m.Push("key", k)
		m.Push("count", i)
	}
	m.Sort("count", "int_r")
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
		m.Push(key, val)
		_hash_delete(m, prefix, chain, kit.MDB_HASH, key)
	})
}
func _hash_search(m *ice.Message, prefix, key, field, value string) {
	m.Richs(prefix, key, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		if field != "" && value != val[field] {
			return
		}
		m.Push(key, value)
	})
}

func _list_insert(m *ice.Message, prefix, key string, arg ...string) int {
	m.Log_INSERT("prefix", prefix, arg[0], arg[1])
	return m.Grow(prefix, key, kit.Dict(arg))
}
func _list_delete(m *ice.Message, prefix, chain, field, value string) {
}
func _list_select(m *ice.Message, prefix, key, field, value string) {
	fields := strings.Split(kit.Select("time,type,name,text", m.Option(FIELDS)), ",")
	m.Grows(prefix, key, field, value, func(index int, val map[string]interface{}) {
		if val[kit.MDB_META] != nil {
			val = val[kit.MDB_META].(map[string]interface{})
		}

		if m.Option(FIELDS) == "detail" {
			m.Push("detail", val)
		} else {
			m.Push(key, val, fields, val[kit.MDB_META])
		}
	})
	if m.Option(FIELDS) != "detail" {
		m.Sort(kit.MDB_ID, "int_r")
	}
}
func _list_modify(m *ice.Message, prefix, key string, field, value string, arg ...string) {
	m.Grows(prefix, key, field, value, func(index int, value map[string]interface{}) {
		if value[kit.MDB_META] != nil {
			value = value[kit.MDB_META].(map[string]interface{})
		}
		for i := 0; i < len(arg)-1; i += 2 {
			if arg[i] == field {
				continue
			}
			kit.Value(value, arg[i], arg[i+1])
		}
		m.Log_MODIFY("prefix", prefix, field, value, kit.Format(arg))
	})
}
func _list_import(m *ice.Message, prefix, key, file string) {
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
		m.Grow(prefix, key, data)
		count++
	}
	m.Log_IMPORT(kit.MDB_KEY, kit.Keys(prefix, key), kit.MDB_COUNT, count)
	m.Echo(kit.Keys(file, CSV))
}
func _list_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	count := 0
	head := []string{}
	m.Grows(prefix, key, "", "", func(index int, value map[string]interface{}) {
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
func _list_search(m *ice.Message, prefix, key, field, value string) {
	list := []interface{}{}
	files := map[string]bool{}
	m.Richs(prefix, key, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		kit.Fetch(kit.Value(value, "meta.record"), func(index int, value map[string]interface{}) {
			file := value["file"].(string)
			if _, ok := files[file]; ok {
				list = append(list, file)
			} else {
				files[file] = true
			}
		})
	})
	defer m.Cost("search")

	task.Wait(list, func(task *task.Task, lock *task.Lock) error {
		kit.CSV(kit.Format(task.Arg), 100000, func(index int, line map[string]string, head []string) {
			if line[field] != value {
				return
			}

			defer lock.WLock()()
			m.Push("", line)
		})
		return nil
	})
}

const (
	ErrDenyModify = "deny modify "
)
const (
	CSV  = "csv"
	JSON = "json"
)
const (
	DICT = "dict"
	META = "meta"
	HASH = "hash"
	LIST = "list"
)
const (
	FIELDS = "fields"
	DETAIL = "detail"
	CREATE = "create"
	RENAME = "rename"
	REMOVE = "remove"
	COMMIT = "commit"
	INVITE = "invite"

	INSERT = "insert"
	DELETE = "delete"
	SELECT = "select"
	MODIFY = "modify"

	IMPORT = "import"
	EXPORT = "export"
	PRUNES = "prunes"
	INPUTS = "inputs"
	SCRIPT = "script"
)

var Index = &ice.Context{Name: "mdb", Help: "数据模块",
	Commands: map[string]*ice.Command{
		INSERT: {Name: "insert conf key type arg...", Help: "添加", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				m.Echo(_hash_insert(m, arg[0], arg[1], arg[3:]...))
			case LIST:
				m.Echo("%d", _list_insert(m, arg[0], arg[1], arg[3:]...))
			}
		}},
		DELETE: {Name: "delete conf key type field value arg...", Help: "删除", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				_hash_delete(m, arg[0], arg[1], arg[3], arg[4])
			case LIST:
				_list_delete(m, arg[0], arg[1], arg[3], arg[4])
			}
		}},
		SELECT: {Name: "select conf key type field value", Help: "数据查询", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case HASH:
				_hash_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select(kit.MDB_FOREACH, arg, 4))
			case LIST:
				_list_select(m, arg[0], arg[1], kit.Select("", arg, 3), kit.Select("", arg, 4))
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
		IMPORT: {Name: "import conf key type file", Help: "导入数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch file := _file_name(m, arg...); arg[2] {
			case HASH:
				_hash_import(m, arg[0], arg[1], file)
			case LIST:
				_list_import(m, arg[0], arg[1], file)
			}
		}},
		EXPORT: {Name: "export conf key type [name]", Help: "导出数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch file := _file_name(m, arg...); arg[2] {
			case HASH:
				_hash_export(m, arg[0], arg[1], file)
			case LIST:
				_list_export(m, arg[0], arg[1], file)
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
				_hash_inputs(m, arg[0], arg[1], arg[3], kit.Select("", arg, 4))
			case LIST:
			}
		}},
	},
}

func init() {
	ice.Index.Register(Index, nil,
		INSERT, DELETE, SELECT, MODIFY,
		IMPORT, EXPORT, PRUNES, INPUTS,
		PLUGIN, RENDER, SEARCH, ENGINE,
	)
}
