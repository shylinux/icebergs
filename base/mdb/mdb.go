package mdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/task"

	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
)

func _list_import(m *ice.Message, prefix, key, file string) {
	f, e := os.Open(file)
	m.Assert(e)
	defer m.Cmdy(web.STORY, web.CATCH, CSV, file)
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
}
func _list_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(kit.Keys(file, CSV))
	m.Assert(e)
	defer m.Cmdy(web.STORY, web.CATCH, CSV, p)
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
}
func _list_select(m *ice.Message, prefix, key, limit, offend, field, value string) {
	m.Option("cache.limit", limit)
	m.Option("cache.offend", offend)
	m.Grows(prefix, key, field, value, func(index int, value map[string]interface{}) {
		m.Push("", value)
	})
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

	task.Sync(list, func(task *task.Task, lock *task.Lock) error {
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

func _hash_search(m *ice.Message, prefix, key, field, value string) {
	m.Richs(prefix, key, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		if field != "" && value != val[field] {
			return
		}
		m.Push(key, value)
	})
}
func _hash_import(m *ice.Message, prefix, key, file string) {
	f, e := os.Open(file)
	m.Assert(e)
	defer m.Cmdy(web.STORY, web.CATCH, JSON, file)
	defer f.Close()

	list := map[string]interface{}{}
	de := json.NewDecoder(f)
	de.Decode(&list)

	count := 0
	for _, data := range list {
		// 导入数据
		m.Rich(prefix, key, data)
		count++
	}
	m.Log_IMPORT(kit.MDB_KEY, kit.Keys(prefix, key), kit.MDB_COUNT, count)
}
func _hash_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(kit.Keys(file, JSON))
	m.Assert(e)
	defer m.Cmdy(web.STORY, web.CATCH, JSON, p)
	defer f.Close()

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	en.Encode(m.Confv(prefix, kit.Keys(key, HASH)))
	m.Log_EXPORT(kit.MDB_FILE, p)
}
func _dict_import(m *ice.Message, prefix, key, file string) {
	f, e := os.Open(file)
	m.Assert(e)
	m.Cmdy(web.STORY, web.CATCH, JSON, file)
	defer f.Close()

	data := map[string]interface{}{}
	de := json.NewDecoder(f)
	de.Decode(&data)

	count := 0
	for k, v := range data {
		m.Log_MODIFY(kit.MDB_KEY, kit.Keys(prefix, key), "k", k, "v", v)
		m.Conf(prefix, kit.Keys(key, k), v)
		count++
	}
	m.Log_EXPORT(kit.MDB_FILE, file, kit.MDB_COUNT, count)
}
func _dict_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(kit.Keys(file, JSON))
	m.Assert(e)
	defer m.Cmdy(web.STORY, web.CATCH, JSON, p)
	defer f.Close()

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	en.Encode(m.Confv(prefix, kit.Keys(key)))
	m.Log_EXPORT(kit.MDB_FILE, p)
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
)
const (
	IMPORT = "import"
	EXPORT = "export"
	SELECT = "select"
	SEARCH = "search"
)

var Index = &ice.Context{Name: "mdb", Help: "数据模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		IMPORT: {Name: "import conf key type file", Help: "导入数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case LIST:
				_list_import(m, arg[0], arg[1], arg[3])
			case HASH:
				_hash_import(m, arg[0], arg[1], arg[3])
			case DICT:
				_dict_import(m, arg[0], arg[1], arg[3])
			}
		}},
		EXPORT: {Name: "export conf key type [name]", Help: "导出数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch file := kit.Select(kit.Select(arg[0], arg[0]+":"+arg[1], arg[1] != ""), arg, 3); arg[2] {
			case LIST:
				_list_export(m, arg[0], arg[1], file)
			case HASH:
				_hash_export(m, arg[0], arg[1], file)
			case DICT:
				_dict_export(m, arg[0], arg[1], file)
			}
		}},

		SELECT: {Name: "select conf key type [limit [offend [key value]]]", Help: "数据查询", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case LIST:
				_list_select(m, arg[0], arg[1], kit.Select("10", arg, 3), kit.Select("0", arg, 4), kit.Select("", arg, 5), kit.Select("", arg, 6))
			}
		}},
		SEARCH: {Name: "search conf key type key value", Help: "数据查询", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case LIST:
				_list_search(m, arg[0], arg[1], arg[3], arg[4])
			case HASH:
				_hash_search(m, arg[0], arg[1], arg[3], arg[4])
			}
		}},
	},
}

func init() {
	ice.Index.Register(Index, nil, IMPORT, EXPORT, SELECT, SEARCH)
}
