package mdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

func _list_import(m *ice.Message, prefix, key, file string) {
	f, e := os.Open(file)
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
}
func _list_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(kit.Keys(file, "csv"))
	m.Assert(e)
	defer f.Close()
	defer m.Cmdy(web.STORY, "catch", "csv", p)

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
func _hash_import(m *ice.Message, prefix, key, file string) {
	f, e := os.Open(file)
	m.Assert(e)
	defer f.Close()

	defer m.Log_IMPORT(kit.MDB_FILE, file)

	data := map[string]interface{}{}
	de := json.NewDecoder(f)
	de.Decode(&data)

	for k, v := range data {
		m.Log_MODIFY(kit.MDB_KEY, kit.Keys(prefix, key, kit.MDB_HASH), "k", k, "v", v)
		m.Conf(prefix, kit.Keys(key, k), v)
	}
}
func _hash_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(kit.Keys(file, "json"))
	m.Assert(e)
	defer f.Close()
	defer m.Cmdy(web.STORY, "catch", "json", p)

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	en.Encode(m.Confv(prefix, kit.Keys(key, kit.MDB_HASH)))
	m.Log_EXPORT(kit.MDB_FILE, p)
}
func _dict_import(m *ice.Message, prefix, key, file string) {
	f, e := os.Open(file)
	m.Assert(e)
	defer f.Close()

	defer m.Log_IMPORT(kit.MDB_FILE, file)

	data := map[string]interface{}{}
	de := json.NewDecoder(f)
	de.Decode(&data)

	for k, v := range data {
		m.Log_MODIFY(kit.MDB_KEY, kit.Keys(prefix, key), "k", k, "v", v)
		m.Conf(prefix, kit.Keys(key, k), v)
	}
}
func _dict_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(kit.Keys(file, "json"))
	m.Assert(e)
	defer f.Close()
	defer m.Cmdy(web.STORY, "catch", "json", p)

	en := json.NewEncoder(f)
	en.SetIndent("", "  ")
	en.Encode(m.Confv(prefix, kit.Keys(key)))
	m.Log_EXPORT(kit.MDB_FILE, p)
}

const IMPORT = "import"
const EXPORT = "export"

var Index = &ice.Context{Name: "mdb", Help: "数据模块",
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		kit.MDB_IMPORT: {Name: "import conf key type file", Help: "导入数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case kit.MDB_LIST:
				_list_import(m, arg[0], arg[1], arg[3])
			case kit.MDB_HASH:
				_hash_import(m, arg[0], arg[1], arg[3])
			case kit.MDB_DICT:
				_dict_import(m, arg[0], arg[1], arg[3])
			}
		}},
		kit.MDB_EXPORT: {Name: "export conf key type [name]", Help: "导出数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch file := kit.Select(kit.Select(arg[0], arg[0]+":"+arg[1], arg[1] != ""), arg, 3); arg[2] {
			case kit.MDB_LIST:
				_list_export(m, arg[0], arg[1], file)
			case kit.MDB_HASH:
				_hash_export(m, arg[0], arg[1], file)
			case kit.MDB_DICT:
				_dict_export(m, arg[0], arg[1], file)
			}
		}},

		"update": {Name: "update config table index key value", Help: "修改数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			meta := m.Confm(arg[0], arg[1]+".meta")
			index := kit.Int(arg[2]) - kit.Int(meta["offset"]) - 1

			data := m.Confm(arg[0], arg[1]+".list."+kit.Format(index))
			for i := 3; i < len(arg)-1; i += 2 {
				kit.Value(data, arg[i], arg[i+1])
			}
		}},
		"select": {Name: "select config table index offend limit match value", Help: "修改数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 3 {
				meta := m.Confm(arg[0], arg[1]+".meta")
				index := kit.Int(arg[2]) - kit.Int(meta["offset"]) - 1

				data := m.Confm(arg[0], arg[1]+".list."+kit.Format(index))
				m.Push(arg[2], data)
			} else {
				m.Option("cache.offend", kit.Select("0", arg, 3))
				m.Option("cache.limit", kit.Select("10", arg, 4))
				fields := strings.Split(arg[7], " ")
				m.Grows(arg[0], arg[1], kit.Select("", arg, 5), kit.Select("", arg, 6), func(index int, value map[string]interface{}) {
					m.Push("id", value, fields)
				})
			}
		}},

		ice.MDB_DELETE: {Name: "delete conf key type", Help: "删除", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[2] {
			case kit.MDB_DICT:
				m.Log(ice.LOG_DELETE, "%s: %s", arg[1], m.Conf(arg[0], arg[1]))
				m.Echo("%s", m.Conf(arg[0], arg[1]))
				m.Conf(arg[0], arg[1], "")
			case kit.MDB_META:
			case kit.MDB_LIST:
			case kit.MDB_HASH:
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil, IMPORT, EXPORT) }
