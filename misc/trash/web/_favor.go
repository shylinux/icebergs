package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"

	"encoding/csv"
	"os"
	"strings"
)

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
			m.SortIntR(kit.MDB_ID)
			return
		}
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			// 详细信息
			m.Push("detail", value)
			m.Optionv("value", value)
			m.Push(kit.MDB_KEY, mdb.RENDER)
			m.Push(kit.MDB_VALUE, m.Cmdx(m.Conf(FAVOR, kit.Keys(kit.MDB_META, mdb.RENDER, value[kit.MDB_TYPE]))))
		})
	})
}
func _favor_show(m *ice.Message, kind string, name, text interface{}, arg ...string) {
	if kind == "" && name == "" {
		m.Richs(FAVOR, nil, m.Option(kit.MDB_ZONE), func(key string, val map[string]interface{}) {
			m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, m.Option(kit.MDB_ID), func(index int, value map[string]interface{}) {
				kind = kit.Format(value[kit.MDB_TYPE])
				name = kit.Format(value[kit.MDB_NAME])
				text = kit.Format(value[kit.MDB_TEXT])
				arg = kit.Simple(value[kit.MDB_EXTRA])
				m.Log_EXPORT(kit.MDB_META, FAVOR, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
			})
		})
	}

	if cmd := m.Conf(FAVOR, kit.Keys("meta.render", kind)); cmd != "" {
		m.Cmdy(cmd, kind, name, text, arg)
		return
	}
	m.Cmdy(kind, "action", "show", kind, name, text, arg)
}
func _favor_sync(m *ice.Message, zone, route, favor string, arg ...string) {
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		val = val[kit.MDB_META].(map[string]interface{})
		remote := kit.Keys("remote", route, favor)
		count := kit.Int(kit.Value(val, kit.Keys(kit.MDB_COUNT)))

		pull := kit.Int(kit.Value(val, kit.Keys(remote, kit.MDB_PULL)))
		m.Cmd(SPIDE, route, "msg", "/favor/pull", FAVOR, favor, "begin", pull+1).Table(func(index int, value map[string]string, head []string) {
			_favor_insert(m, favor, value[kit.MDB_TYPE], value[kit.MDB_NAME], value[kit.MDB_TEXT], value[kit.MDB_EXTRA])
			pull = kit.Int(value[kit.MDB_ID])
		})

		m.Option("cache.limit", count-kit.Int(kit.Value(val, kit.Keys(remote, kit.MDB_PUSH))))
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			m.Cmd(SPIDE, route, "msg", "/favor/push", FAVOR, favor,
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
				m.Cmdy(PROXY, p, FAVOR, zone, kit.MDB_TYPE, value[kit.MDB_TYPE],
					kit.MDB_NAME, value[kit.MDB_NAME], kit.MDB_TEXT, value[kit.MDB_TEXT],
					kit.Format(value[kit.MDB_EXTRA]))
			})
		})
	}
}
func _favor_share(m *ice.Message, zone, id string, arg ...string) {
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Cmdy(SHARE, value[kit.MDB_TYPE], value[kit.MDB_NAME], value[kit.MDB_TYPE], kit.Format(value[kit.MDB_EXTRA]))
		})
	})
}
func _favor_commit(m *ice.Message, zone, id string, arg ...string) {
	m.Echo("list: ")
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Cmdy(STORY, "add", value[kit.MDB_TYPE], value[kit.MDB_NAME], value[kit.MDB_TEXT])
		})
	})
}
func _favor_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(FAVOR, nil, kit.Select(kit.MDB_FOREACH, ""), func(key string, val map[string]interface{}) {
		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			if name != value[kit.MDB_NAME] && !strings.Contains(kit.Format(value[kit.MDB_TEXT]), name) {
				return
			}
			m.Push("pod", m.Option(ice.MSG_USERPOD))
			m.Push("ctx", m.Prefix())
			m.Push("cmd", FAVOR)
			m.Push(key, value, []string{kit.MDB_TIME}, val)
			m.Push(kit.MDB_SIZE, kit.FmtSize(int64(len(kit.Format(value[kit.MDB_TEXT])))))
			m.Push(key, value, []string{kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT}, val)
		})
	})
}
func _favor_render(m *ice.Message, kind, name, text string, arg ...string) {
}

func _favor_modify(m *ice.Message, zone, id, pro, set, old string) {
	m.Richs(FAVOR, nil, zone, func(key string, val map[string]interface{}) {
		switch pro {
		case kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TIME:
			m.Warn(true, ice.ErrNotRight, pro)
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
func _favor_insert(m *ice.Message, zone, kind, name interface{}, text interface{}, arg ...string) {
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

func FavorInsert(m *ice.Message, zone, kind string, name interface{}, text interface{}, extra ...string) {
	_favor_create(m, zone)
	_favor_insert(m, zone, kind, name, text, extra...)
}
func FavorList(m *ice.Message, favor, id string, fields ...string) *ice.Message {
	_favor_list(m, favor, id, fields...)
	return m
}
func FavorShow(m *ice.Message, kind string, name, text interface{}, arg ...string) *ice.Message {
	_favor_show(m, kind, name, text, arg...)
	return m
}

const (
	_EXPORT = "usr/local/export/web.favor/favor.csv"
)

const FAVOR = "favor"

const PLUGIN = "plugin"

const ( // TYPE
	TYPE_RIVER  = "river"
	TYPE_STORM  = "storm"
	TYPE_ACTION = "action"
	TYPE_ACTIVE = "active"

	TYPE_DRIVE = "drive"
	TYPE_SHELL = "shell"
	TYPE_VIMRC = "vimrc"
	TYPE_TABLE = "table"
	TYPE_INNER = "inner"
	TYPE_MEDIA = "media"
)

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
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					_favor_export(m, kit.Select(_EXPORT, arg, 0))
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_favor_import(m, kit.Select(_EXPORT, arg, 0))
				}},
				mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_favor_create(m, arg[0])
				}},
				mdb.INSERT: {Name: "insert zone type name text", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
					_favor_create(m, arg[0])
					_favor_insert(m, arg[0], arg[1], arg[2], kit.Select("", arg, 3))
				}},
				mdb.MODIFY: {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_favor_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1], kit.Select("", arg, 2))
				}},
				mdb.COMMIT: {Name: "commit arg...", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
					_favor_commit(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg...)
				}},
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_favor_search(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},
				mdb.RENDER: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_favor_render(m, arg[0], arg[1], arg[2], arg[3:]...)
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
				kit.MDB_SHOW: {Name: "show type name text arg...", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) > 2 {
						_favor_show(m, arg[0], arg[1], arg[2], arg[3:]...)
					} else {
						_favor_show(m, "", "", "")
					}
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
			"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(SHARE, nil, kit.Select(m.Option(kit.SSH_SHARE), arg, 0), func(key string, value map[string]interface{}) {
					m.Log_SELECT(kit.MDB_META, SHARE, "arg", arg, "value", kit.Format(value))
					if m.Warn(m.Option(ice.MSG_USERROLE) != aaa.ROOT && kit.Time(kit.Format(value[kit.MDB_TIME])) < kit.Time(m.Time()), ice.ErrExpire, arg) {
						m.Echo("expired")
						return
					}

					switch value[kit.MDB_TYPE] {
					case STORY:
						value = _share_story(m, value, arg...)
					}

					if _share_show(m, key, value, kit.Select("", arg, 1), kit.Select("", arg, 2)) {
						return
					}

					switch value[kit.MDB_TYPE] {
					case TYPE_RIVER:
						// 共享群组
						m.Render("redirect", "/", "share", key, "river", kit.Format(value["text"]))

					case TYPE_STORM:
						// 共享应用
						m.Render("redirect", "/", "share", key, "storm", kit.Format(value["text"]), "river", kit.Format(kit.Value(value, "extra.river")))

					case TYPE_ACTION:
						_share_action(m, value, arg...)

					default:
						// 查看数据
						m.Option(kit.MDB_VALUE, value)
						m.Option(kit.MDB_TYPE, value[kit.MDB_TYPE])
						m.Option(kit.MDB_NAME, value[kit.MDB_NAME])
						m.Option(kit.MDB_TEXT, value[kit.MDB_TEXT])
						m.RenderTemplate(m.Conf(SHARE, "meta.template.simple"))
						m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
					}
				})
			}},
		}}, nil)
}
