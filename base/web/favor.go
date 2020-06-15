package web

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"encoding/csv"
	"os"
)

const (
	EXPORT = "usr/export/web.favor/favor.csv"
)

var FAVOR = ice.Name("favor", Index)

func _favor_list(m *ice.Message, zone, id string, fields ...string) {
	m.Richs(FAVOR, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		if val = val[kit.MDB_META].(map[string]interface{}); zone == "" {
			m.Push("", val, []string{
				// 汇总信息
				kit.MDB_TIME, kit.MDB_COUNT, kit.MDB_ZONE,
			})
			m.Sort(kit.MDB_ZONE)
			return
		}
		if zone = kit.Format(kit.Value(val, kit.MDB_ZONE)); id == "" {
			m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				// 列表信息
				m.Push(zone, value, fields, val)
			})
			m.Sort(kit.MDB_ID, "int_r")
			return
		}
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			// 详细信息
			m.Push("detail", value)
			m.Optionv("value", value)
			m.Push(kit.MDB_KEY, kit.MDB_RENDER)
			m.Push(kit.MDB_VALUE, m.Cmdx(m.Conf(ice.WEB_FAVOR, kit.Keys(kit.MDB_META, kit.MDB_RENDER, value[kit.MDB_TYPE]))))
		})
	})
}
func _favor_show(m *ice.Message, zone, id string, arg ...string) {
}
func _favor_sync(m *ice.Message, zone, route, favor string, arg ...string) {
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		val = val[kit.MDB_META].(map[string]interface{})
		remote := kit.Keys("remote", route, favor)
		count := kit.Int(kit.Value(val, kit.Keys(kit.MDB_COUNT)))

		pull := kit.Int(kit.Value(val, kit.Keys(remote, kit.MDB_PULL)))
		m.Cmd(ice.WEB_SPIDE, route, "msg", "/favor/pull", FAVOR, favor, "begin", pull+1).Table(func(index int, value map[string]string, head []string) {
			_favor_insert(m, favor, value[kit.MDB_TYPE], value[kit.MDB_NAME], value[kit.MDB_TEXT], value[kit.MDB_EXTRA])
			pull = kit.Int(value[kit.MDB_ID])
		})

		m.Option("cache.limit", count-kit.Int(kit.Value(val, kit.Keys(remote, kit.MDB_PUSH))))
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			m.Cmd(ice.WEB_SPIDE, route, "msg", "/favor/push", FAVOR, favor,
				kit.MDB_TYPE, value[kit.MDB_TYPE], kit.MDB_NAME, value[kit.MDB_NAME], kit.MDB_TEXT, value[kit.MDB_TEXT],
				kit.MDB_EXTRA, kit.Format(value[kit.MDB_EXTRA]),
			)
			pull++
		})
		kit.Value(val, kit.Keys(remote, kit.MDB_PULL), pull)
		kit.Value(val, kit.Keys(remote, kit.MDB_PUSH), kit.Value(val, kit.MDB_COUNT))
		m.Echo("%d", kit.Value(val, kit.MDB_COUNT))
		return
	})
}
func _favor_pull(m *ice.Message, zone string) {
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		m.Option("cache.limit", kit.Int(kit.Value(val, "meta.count"))+1-kit.Int(m.Option("begin")))
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			m.Log_EXPORT(kit.MDB_VALUE, value)
			m.Push(key, value, []string{kit.MDB_ID, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
			m.Push(kit.MDB_EXTRA, kit.Format(value[kit.MDB_EXTRA]))
		})
	})
}
func _favor_push(m *ice.Message, zone, id string, arg ...string) {
}
func _favor_proxy(m *ice.Message, zone, id string, arg ...string) {
	if p := kit.Select(m.Conf(FAVOR, kit.Keys(kit.MDB_META, kit.MDB_PROXY)), m.Option("you")); p != "" {
		m.Option("you", "")
		// 分发数据
		m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
			m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
				m.Cmdy(ice.WEB_PROXY, p, ice.WEB_FAVOR, zone, kit.MDB_TYPE, value[kit.MDB_TYPE],
					kit.MDB_NAME, value[kit.MDB_NAME], kit.MDB_TEXT, value[kit.MDB_TEXT],
					kit.Format(value[kit.MDB_EXTRA]))
			})
		})
	}
}
func _favor_share(m *ice.Message, zone, id string, arg ...string) {
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Cmdy(ice.WEB_SHARE, value[kit.MDB_TYPE], value[kit.MDB_NAME], value[kit.MDB_TYPE], kit.Format(value[kit.MDB_EXTRA]))
		})
	})
}
func _favor_commit(m *ice.Message, zone, id string, arg ...string) {
	m.Echo("list: ")
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Cmdy(ice.WEB_STORY, "add", value[kit.MDB_TYPE], value[kit.MDB_NAME], value[kit.MDB_TEXT])
		})
	})
}

func _favor_modify(m *ice.Message, zone, id, pro, set, old string) {
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		switch pro {
		case kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TIME:
			m.Warn(true, "deny modify %v", pro)
			return
		}

		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			// 修改信息
			m.Log_MODIFY(kit.MDB_META, FAVOR, kit.MDB_ZONE, zone,
				kit.MDB_ID, id, kit.MDB_KEY, pro, kit.MDB_VALUE, set, "old", kit.Value(value, pro))
			kit.Value(value, pro, set)
		})
	})
}
func _favor_insert(m *ice.Message, zone, kind, name, text string, arg ...string) {
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		kit.Value(val, kit.Keys(kit.MDB_META, kit.MDB_TIME), m.Time())

		id := m.Grow(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.Dict(
			kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
			kit.MDB_EXTRA, kit.Dict(arg),
		))
		m.Log_INSERT(kit.MDB_META, FAVOR, kit.MDB_ZONE, zone,
			kit.MDB_ID, id, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
		m.Echo("%d", id)
	})
}
func _favor_create(m *ice.Message, zone string, arg ...string) {
	if m.Richs(FAVOR, nil, zone, nil) == nil {
		m.Rich(FAVOR, nil, kit.Data(kit.MDB_ZONE, zone, arg))
		m.Log_CREATE(kit.MDB_META, FAVOR, kit.MDB_ZONE, zone)
	}
}
func _favor_import(m *ice.Message, file string) {
	f, e := os.Open(file)
	m.Assert(e)
	defer f.Close()

	r := csv.NewReader(f)
	heads, _ := r.Read()

	count := 0
	for {
		lines, e := r.Read()
		if e != nil {
			break
		}

		zone := ""
		data := kit.Dict()
		for i, k := range heads {
			switch k {
			case kit.MDB_ZONE:
				zone = lines[i]
			case kit.MDB_ID:
				continue
			case kit.MDB_EXTRA:
				kit.Value(data, k, kit.UnMarshal(lines[i]))
			default:
				kit.Value(data, k, lines[i])
			}
		}

		_favor_create(m, zone)
		m.Richs(FAVOR, nil, zone, func(key string, value map[string]interface{}) {
			id := m.Grow(FAVOR, kit.Keys(kit.MDB_HASH, key), data)
			m.Log_INSERT(kit.MDB_META, FAVOR, kit.MDB_ZONE, zone, kit.MDB_ID, id, kit.MDB_TYPE, data[kit.MDB_TYPE], kit.MDB_NAME, data[kit.MDB_NAME])
			count++
		})
	}
	m.Log_IMPORT(kit.MDB_FILE, file, kit.MDB_COUNT, count)
}
func _favor_export(m *ice.Message, file string) {
	f, p, e := kit.Create(file)
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	m.Assert(w.Write([]string{
		kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TIME,
		kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT,
		kit.MDB_EXTRA,
	}))

	count := 0
	m.Option("cache.limit", -2)
	m.Richs(FAVOR, nil, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		val = val[kit.MDB_META].(map[string]interface{})
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			m.Assert(w.Write(kit.Simple(
				kit.Format(val[kit.MDB_ZONE]),
				kit.Format(value[kit.MDB_ID]),
				kit.Format(value[kit.MDB_TIME]),
				kit.Format(value[kit.MDB_TYPE]),
				kit.Format(value[kit.MDB_NAME]),
				kit.Format(value[kit.MDB_TEXT]),
				kit.Format(value[kit.MDB_EXTRA]),
			)))
			count++
		})
	})
	m.Log_EXPORT(kit.MDB_FILE, p, kit.MDB_COUNT, count)
}

func FavorInsert(m *ice.Message, zone, kind, name, text string, extra ...string) {
	_favor_create(m, zone)
	_favor_insert(m, zone, kind, name, text, extra...)
}
func FavorList(m *ice.Message, favor, id string, fields ...string) {
	_favor_list(m, favor, id, fields...)
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FAVOR: {Name: "favor", Help: "收藏夹", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_ZONE, "template", favor_template, "proxy", "",
			)},
		},
		Commands: map[string]*ice.Command{
			FAVOR: {Name: "favor zone=auto id=auto auto", Help: "收藏夹", Meta: kit.Dict(
				"detail", []string{"编辑", "收藏", "收录", "导出", "删除"},
			), Action: map[string]*ice.Action{
				kit.MDB_EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					_favor_export(m, kit.Select(EXPORT, arg, 0))
				}},
				kit.MDB_IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_favor_import(m, kit.Select(EXPORT, arg, 0))
				}},
				kit.MDB_CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_favor_create(m, arg[0])
				}},
				kit.MDB_INSERT: {Name: "insert zone type name text", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
					_favor_insert(m, arg[0], arg[1], arg[2], kit.Select("", arg, 3))
				}},
				kit.MDB_MODIFY: {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_favor_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1], kit.Select("", arg, 2))
				}},
				kit.MDB_COMMIT: {Name: "commit arg...", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
					_favor_commit(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg...)
				}},
				kit.MDB_SHARE: {Name: "share arg...", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					_favor_share(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg...)
				}},
				kit.MDB_PROXY: {Name: "proxy arg...", Help: "代理", Hand: func(m *ice.Message, arg ...string) {
					_favor_proxy(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg...)
				}},
				kit.MDB_SYNC: {Name: "sync route favor", Help: "同步", Hand: func(m *ice.Message, arg ...string) {
					_favor_sync(m, m.Option(kit.MDB_ZONE), arg[0], arg[1], arg[2:]...)
				}},
				kit.MDB_SHOW: {Name: "show arg...", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					_favor_show(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				fields := []string{kit.MDB_TIME, kit.MDB_ID, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT}
				if len(arg) > 1 && arg[1] == "extra" {
					fields, arg = append(fields, arg[2:]...), arg[:1]
				}

				if len(arg) < 3 {
					_favor_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1), fields...)
					return
				}

				_favor_create(m, arg[0])
				_favor_insert(m, arg[0], arg[1], arg[2], arg[3], arg[4:]...)
				return
			}},
			"/favor/": {Name: "/favor/", Help: "收藏夹", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case kit.MDB_PULL:
					_favor_pull(m, m.Option(FAVOR))
				case kit.MDB_PUSH:
					_favor_insert(m, m.Option(FAVOR), m.Option(kit.MDB_TYPE),
						m.Option(kit.MDB_NAME), m.Option(kit.MDB_TEXT), m.Option(kit.MDB_EXTRA))
				}
			}},
		}}, nil)
}
