package web

import (
	"encoding/csv"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"os"
)

var FAVOR = ice.Name("favor", Index)

func _favor_list(m *ice.Message, favor, id string, fields ...string) {
	if favor == "" {
		m.Richs(FAVOR, nil, "*", func(key string, value map[string]interface{}) {
			m.Push(key, value["meta"], []string{"time", "count"})
			m.Push("render", kit.Select("spide", kit.Value(value, "meta.render")))
			m.Push(FAVOR, kit.Value(value, "meta.name"))
		})
		m.Sort(FAVOR)
		return
	}

	m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
		if id == "" {
			m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				m.Push("", value, fields)
			})
			return
		}

		m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", id, func(index int, value map[string]interface{}) {
			m.Push("detail", value)
			m.Optionv("value", value)
			m.Push("key", "render")
			m.Push("value", m.Cmdx(m.Conf(ice.WEB_FAVOR, kit.Keys("meta.render", value["type"]))))
		})
	})
}
func _favor_create(m *ice.Message, name string) string {
	favor := m.Rich(ice.WEB_FAVOR, nil, kit.Data(kit.MDB_NAME, name))
	m.Log_CREATE("favor", favor, "name", favor)
	return favor
}
func _favor_insert(m *ice.Message, favor, kind, name, text string, extra ...string) {
	index := m.Grow(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, favor), kit.Dict(
		kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
		"extra", kit.Dict(extra),
	))
	m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
		kit.Value(value, "meta.time", m.Time())
	})
	m.Log_INSERT("favor", favor, "index", index, "name", name, "text", text)
	m.Echo("%d", index)
}
func _favor_modify(m *ice.Message, favor, id, pro, set, old string) {
	m.Richs(FAVOR, nil, favor, func(key string, val map[string]interface{}) {
		switch pro {
		case FAVOR, kit.MDB_ID, kit.MDB_TIME:
			m.Warn(true, "deny modify %v", pro)
			return
		}

		m.Grows(FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			// 修改信息
			m.Log_MODIFY(FAVOR, favor, kit.MDB_ID, id, kit.MDB_KEY, pro, kit.MDB_VALUE, set, "old", old)
			kit.Value(value, pro, set)
		})
	})
}

func FavorInsert(m *ice.Message, favor, kind, name, text string, extra ...string) {
	_favor_insert(m, favor, kind, name, text, extra...)
}
func FavorList(m *ice.Message, favor, id string, fields ...string) {
	_favor_list(m, favor, id, fields...)
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_FAVOR: {Name: "favor", Help: "收藏夹", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, "template", favor_template,
				"proxy", "",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_FAVOR: {Name: "favor favor=auto id=auto auto", Help: "收藏夹", Meta: kit.Dict(
				"exports", []string{"hot", "favor"}, "detail", []string{"编辑", "收藏", "收录", "导出", "删除"},
			), Action: map[string]*ice.Action{
				kit.MDB_MODIFY: {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_favor_modify(m, m.Option(FAVOR), m.Option(kit.MDB_ID), arg[0], arg[1], kit.Select("", arg, 2))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 && arg[0] == "action" {
					favor, id := m.Option("favor"), m.Option("id")
					switch arg[2] {
					case "favor":
						favor = arg[3]
					case "id":
						id = arg[3]
					}

					switch arg[1] {
					case "commit", "收录":
						m.Echo("list: ")
						m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
							m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", id, func(index int, value map[string]interface{}) {
								m.Cmdy(ice.WEB_STORY, "add", value["type"], value["name"], value["text"])
							})
						})
					case "export", "导出":
						m.Echo("list: ")
						if favor == "" {
							m.Cmdy(ice.MDB_EXPORT, ice.WEB_FAVOR, kit.MDB_HASH, kit.MDB_HASH, "favor")
						} else {
							m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
								m.Cmdy(ice.MDB_EXPORT, ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_LIST, favor)
							})
						}
					case "import", "导入":
						if favor == "" {
							m.Cmdy(ice.MDB_IMPORT, ice.WEB_FAVOR, kit.MDB_HASH, kit.MDB_HASH, "favor")
						} else {
							m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
								m.Cmdy(ice.MDB_IMPORT, ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_LIST, favor)
							})
						}
					}
					return
				}

				if len(arg) == 0 {
					_favor_list(m, "", "")
					return
				}

				switch arg[0] {
				case "save":
					f, p, e := kit.Create(arg[1])
					m.Assert(e)
					defer f.Close()
					w := csv.NewWriter(f)

					w.Write([]string{"favor", "type", "name", "text", "extra"})

					n := 0
					m.Option("cache.offend", 0)
					m.Option("cache.limit", -2)
					for _, favor := range arg[2:] {
						m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, val map[string]interface{}) {
							if m.Conf(ice.WEB_FAVOR, kit.Keys("meta.skip", kit.Value(val, "meta.name"))) == "true" {
								return
							}
							m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
								w.Write(kit.Simple(kit.Value(val, "meta.name"), value["type"], value["name"], value["text"], kit.Format(value["extra"])))
								n++
							})
						})
					}
					w.Flush()
					m.Echo("%s: %d", p, n)
					return

				case "load":
					f, e := os.Open(arg[1])
					m.Assert(e)
					defer f.Close()
					r := csv.NewReader(f)

					head, e := r.Read()
					m.Assert(e)
					m.Info("head: %v", head)

					for {
						line, e := r.Read()
						if e != nil {
							break
						}
						m.Cmd(ice.WEB_FAVOR, line)
					}
					return

				case "sync":
					m.Richs(ice.WEB_FAVOR, nil, arg[1], func(key string, val map[string]interface{}) {
						remote := kit.Keys("meta.remote", arg[2], arg[3])
						count := kit.Int(kit.Value(val, kit.Keys("meta.count")))

						pull := kit.Int(kit.Value(val, kit.Keys(remote, "pull")))
						m.Cmd(ice.WEB_SPIDE, arg[2], "msg", "/favor/pull", "favor", arg[3], "begin", pull+1).Table(func(index int, value map[string]string, head []string) {
							m.Cmd(ice.WEB_FAVOR, arg[1], value["type"], value["name"], value["text"], value["extra"])
							pull = kit.Int(value["id"])
						})

						m.Option("cache.limit", count-kit.Int(kit.Value(val, kit.Keys(remote, "push"))))
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							m.Cmd(ice.WEB_SPIDE, arg[2], "msg", "/favor/push", "favor", arg[3],
								"type", value["type"], "name", value["name"], "text", value["text"],
								"extra", kit.Format(value["extra"]),
							)
							pull++
						})
						kit.Value(val, kit.Keys(remote, "pull"), pull)
						kit.Value(val, kit.Keys(remote, "push"), kit.Value(val, "meta.count"))
						m.Echo("%d", kit.Value(val, "meta.count"))
						return
					})
					return
				}

				fields := []string{kit.MDB_TIME, kit.MDB_ID, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT}
				if len(arg) > 1 && arg[1] == "extra" {
					fields, arg = append(fields, arg[2:]...), arg[:1]
				}
				m.Option("favor", arg[0])
				if len(arg) < 3 {
					// 收藏列表
					_favor_list(m, arg[0], kit.Select("", arg, 1), fields...)
					return
				}

				favor := ""
				if m.Richs(ice.WEB_FAVOR, nil, arg[0], func(key string, value map[string]interface{}) {
					favor = key
				}) == nil {
					// 创建收藏
					favor = _favor_create(m, arg[0])
				}

				if len(arg) == 3 {
					arg = append(arg, "")
				}
				_favor_insert(m, favor, arg[1], arg[2], arg[3], arg[4:]...)
				return

				// 分发数据
				if p := kit.Select(m.Conf(ice.WEB_FAVOR, "meta.proxy"), m.Option("you")); p != "" {
					m.Option("you", "")
					m.Cmdy(ice.WEB_PROXY, p, ice.WEB_FAVOR, arg)
				}
			}},
			"/favor/": {Name: "/story/", Help: "收藏夹", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

				switch arg[0] {
				case "pull":
					m.Richs(ice.WEB_FAVOR, nil, m.Option("favor"), func(key string, value map[string]interface{}) {
						m.Option("cache.limit", kit.Int(kit.Value(value, "meta.count"))+1-kit.Int(m.Option("begin")))
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							m.Log(ice.LOG_EXPORT, "%v", value)
							m.Push("", value, []string{"id", "type", "name", "text"})
							m.Push("extra", kit.Format(value["extra"]))
						})
					})
				case "push":
					m.Cmdy(ice.WEB_FAVOR, m.Option("favor"), m.Option("type"), m.Option("name"), m.Option("text"), m.Option("extra"))
				}
			}},
		}}, nil)
}
