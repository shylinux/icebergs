package mdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bytes"
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"sort"
	"strings"
)

func distance(lat1, long1, lat2, long2 float64) float64 {
	lat1 = lat1 * math.Pi / 180
	long1 = long1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	long2 = long2 * math.Pi / 180
	return 2 * 6371 * math.Asin(math.Sqrt(math.Pow(math.Sin(math.Abs(lat1-lat2)/2), 2)+math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(math.Abs(long1-long2)/2), 2)))
}

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

		ice.MDB_IMPORT: {Name: "import conf key type file", Help: "导入数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			msg := m.Cmd(ice.WEB_STORY, "index", arg[3])

			switch arg[2] {
			case kit.MDB_DICT:
			case kit.MDB_META:
			case kit.MDB_LIST:
				buf := bytes.NewBufferString(msg.Append("text"))
				r := csv.NewReader(buf)
				if msg.Append("file") != "" {
					if f, e := os.Open(msg.Append("file")); m.Assert(e) {
						// 导入文件
						r = csv.NewReader(f)
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
					n := m.Grow(arg[0], arg[1], data)
					m.Log(ice.LOG_INSERT, "index: %d value: %v", n, data)
				}

			case kit.MDB_HASH:
				data := map[string]interface{}{}
				m.Assert(json.Unmarshal([]byte(msg.Append("text")), &data))
				for k, v := range data {
					m.Conf(arg[0], kit.Keys(arg[1], k), v)
					m.Log(ice.LOG_MODIFY, "%s: %s", k, v)
				}
			}
		}},
		ice.MDB_EXPORT: {Name: "export conf key type name", Help: "导出数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			name := kit.Select(kit.Select(arg[0], arg[0]+":"+arg[1], arg[1] != ""), arg, 3)

			switch arg[2] {
			case kit.MDB_DICT:
			case kit.MDB_META:
			case kit.MDB_LIST:
				f, p, e := kit.Create("var/temp/" + kit.Keys(name, "csv"))
				m.Assert(e)
				w := csv.NewWriter(f)

				head := []string{}
				m.Grows(arg[0], arg[1], "", "", func(index int, value map[string]interface{}) {
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
				w.Flush()

				m.Cmdy(ice.WEB_STORY, "catch", "csv", p)

			case kit.MDB_HASH:
				m.Cmdy(ice.WEB_STORY, "add", "json", name, kit.Formats(m.Confv(arg[0], arg[1])))
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

func init() { ice.Index.Register(Index, nil) }
