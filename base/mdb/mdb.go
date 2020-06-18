package mdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"bytes"
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"path"
	"sort"
	"strings"
)

func _list_import(m *ice.Message, prefix, key, file, text string) {
	buf := bytes.NewBufferString(text)
	r := csv.NewReader(buf)

	if file != "" {
		if f, e := os.Open(file); m.Assert(e) {
			defer f.Close()
			r = csv.NewReader(f)
			// 导入文件
		}
	}

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
		n := m.Grow(prefix, key, data)
		m.Log_INSERT(kit.MDB_KEY, kit.Keys(prefix, key), kit.MDB_ID, n)
	}
}
func _list_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(path.Join("usr/export", kit.Keys(file, "csv")))
	m.Assert(e)
	defer f.Close()
	defer m.Cmdy(web.STORY, "catch", "csv", p)

	w := csv.NewWriter(f)
	defer w.Flush()

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
	})
}
func _hash_import(m *ice.Message, prefix, key, file, text string) {
	data := kit.Parse(nil, "", kit.Split(text, "\t: ,\n")...).(map[string]interface{})
	for k, v := range data {
		m.Log_MODIFY(kit.MDB_KEY, kit.Keys(prefix, key), "k", k, "v", v)
		m.Conf(prefix, kit.Keys(key, k), v)
	}
}
func _hash_export(m *ice.Message, prefix, key, file string) {
	f, p, e := kit.Create(path.Join("usr/export", kit.Keys(file, "json")))
	m.Assert(e)
	defer f.Close()
	defer m.Cmdy(web.STORY, "catch", "json", p)

	en := json.NewEncoder(f)
	en.Encode(m.Confv(prefix, key))
}

func distance(lat1, long1, lat2, long2 float64) float64 {
	lat1 = lat1 * math.Pi / 180
	long1 = long1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	long2 = long2 * math.Pi / 180
	return 2 * 6371 * math.Asin(math.Sqrt(math.Pow(math.Sin(math.Abs(lat1-lat2)/2), 2)+math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(math.Abs(long1-long2)/2), 2)))
}

const IMPORT = "import"
const EXPORT = "export"

var Index = &ice.Context{Name: "mdb", Help: "数据模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"location": {Name: "location", Help: "定位", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(m.Prefix("location"))
		}},
		kit.MDB_IMPORT: {Name: "import conf key type file", Help: "导入数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch msg := m.Cmd(ice.WEB_STORY, "index", arg[3]); arg[2] {
			case kit.MDB_LIST:
				_list_import(m, arg[0], arg[1], msg.Append(kit.MDB_FILE), msg.Append(kit.MDB_TEXT))
			case kit.MDB_HASH:
				_hash_import(m, arg[0], arg[1], msg.Append(kit.MDB_FILE), msg.Append(kit.MDB_TEXT))
			}
		}},
		kit.MDB_EXPORT: {Name: "export conf key type [name]", Help: "导出数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch file := kit.Select(kit.Select(arg[0], arg[0]+":"+arg[1], arg[1] != ""), arg, 3); arg[2] {
			case kit.MDB_LIST:
				_list_export(m, arg[0], arg[1], file)
			case kit.MDB_HASH:
				_hash_export(m, arg[0], arg[1], file)
			}
		}},

		"location": {Name: "location", Help: "location", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Grows("location", nil, "", "", func(index int, value map[string]interface{}) {
					m.Push("", value)
				})
				return
			}
			if len(arg) == 1 {
				m.Richs("location", nil, arg[0], func(key string, value map[string]interface{}) {
					m.Info("what %v", value)
					m.Push("detail", value)
				})
				return
			}
			if len(arg) == 2 {
				m.Richs("aaa.location", nil, "*", func(key string, value map[string]interface{}) {
					m.Push("name", value["name"])
					m.Push("distance", kit.Int(distance(
						float64(kit.Int(arg[0]))/100000,
						float64(kit.Int(arg[1]))/100000,
						float64(kit.Int(value["latitude"]))/100000,
						float64(kit.Int(value["longitude"]))/100000,
					)*1000))
				})
				m.Sort("distance", "int")
				return
			}

			data := m.Richs("location", nil, arg[0], nil)
			if data != nil {
				data["count"] = kit.Int(data["count"]) + 1
			} else {
				data = kit.Dict("name", arg[0], "address", arg[1], "latitude", arg[2], "longitude", arg[3], "count", 1)
				m.Rich("location", nil, data)
			}
			m.Grow("location", nil, data)
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
