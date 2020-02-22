package mdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

var Index = &ice.Context{Name: "mdb", Help: "数据模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
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
		ice.MDB_IMPORT: {Name: "import", Help: "导入数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			msg := m.Cmd(ice.WEB_STORY, "index", arg[3])

			switch arg[2] {
			case kit.MDB_DICT:
			case kit.MDB_META:
			case kit.MDB_LIST:
				buf := bytes.NewBufferString(msg.Append("text"))
				r := csv.NewReader(buf)
				if msg.Append("file") != "" {
					if f, e := os.Open(msg.Append("file")); m.Assert(e) {
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
						data[k] = line[i]
					}
					m.Grow(arg[0], arg[1], data)
					m.Info("import %v", data)
				}

			case kit.MDB_HASH:
				data := map[string]interface{}{}
				m.Assert(json.Unmarshal([]byte(msg.Append("text")), &data))
				for k, v := range data {
					m.Conf(arg[0], kit.Keys(arg[1], k), v)
					m.Info("import %v", v)
				}
			}
		}},
		ice.MDB_EXPORT: {Name: "export conf key list|hash", Help: "导出数据", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			name := kit.Select(kit.Select(arg[0], arg[0]+":"+arg[1], arg[1] != ""), arg, 3)

			switch arg[2] {
			case kit.MDB_DICT:
			case kit.MDB_META:
			case kit.MDB_LIST:
				buf := bytes.NewBuffer(make([]byte, 0, 1024))
				w := csv.NewWriter(buf)
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

				m.Cmdy(ice.WEB_STORY, "add", "csv", name, string(buf.Bytes()))

			case kit.MDB_HASH:
				m.Cmdy(ice.WEB_STORY, "add", "json", name, kit.Formats(m.Confv(arg[0], arg[1])))
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
