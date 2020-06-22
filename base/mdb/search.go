package mdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
	"strings"
	"sync"
)

const SEARCH = "search"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SEARCH: {Name: "search", Help: "搜索引擎", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TYPE)},
		},
		Commands: map[string]*ice.Command{
			SEARCH: {Name: "search word type...", Help: "搜索引擎", Action: map[string]*ice.Action{
				CREATE: {Name: "create type name [text]", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Rich(SEARCH, nil, kit.Dict(kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, kit.Select("", arg, 2)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

				if strings.Contains(arg[0], ";") {
					arg = strings.Split(arg[0], ";")
				}
				if len(arg) > 2 {
					for _, k := range strings.Split(arg[2], ",") {
						m.Richs(SEARCH, nil, k, func(key string, value map[string]interface{}) {
							m.Cmdy(kit.Keys(value[kit.MDB_TEXT], value[kit.MDB_NAME]), SEARCH, value[kit.MDB_TYPE], arg[0], kit.Select("", arg, 1))
						})
					}
				} else {
					m.Richs(SEARCH, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						m.Cmdy(kit.Keys(value[kit.MDB_TEXT], value[kit.MDB_NAME]), SEARCH, value[kit.MDB_TYPE], arg[0], kit.Select("", arg, 1))
					})
				}

				return
				if len(arg) < 2 {
					m.Cmdy("web.label", arg)
					return
				}

				switch arg[0] {
				case "add":
					if m.Richs(cmd, nil, arg[1], nil) == nil {
						m.Rich(cmd, nil, kit.Data(kit.MDB_NAME, arg[1]))
					}
					m.Richs(cmd, nil, arg[1], func(key string, value map[string]interface{}) {
						m.Grow(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(
							kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3:],
						))
					})
				case "get":
					wg := &sync.WaitGroup{}
					m.Richs(cmd, nil, arg[1], func(key string, value map[string]interface{}) {
						wg.Add(1)
						m.Gos(m, func(m *ice.Message) {
							m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
								m.Cmdy(value[kit.MDB_TEXT], arg[2:])
							})
							wg.Done()
						})
					})
					wg.Wait()
				case "set":
					if arg[1] != "" {
						m.Cmdy("web.space", arg[1], "web.chat.search", "set", "", arg[2:])
						break
					}

					m.Richs(cmd, nil, arg[2], func(key string, value map[string]interface{}) {
						m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							m.Cmdy(value[kit.MDB_TEXT], "set", arg[3:])
						})
					})
				default:
					if len(arg) < 4 {
						m.Richs(cmd, nil, kit.Select(kit.MDB_FOREACH, arg, 2), func(key string, val map[string]interface{}) {
							if len(arg) < 3 {
								m.Push(key, val[kit.MDB_META], []string{kit.MDB_TIME, kit.MDB_NAME, kit.MDB_COUNT})
								return
							}
							m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
								m.Push("", value, []string{kit.MDB_TIME})
								m.Push("group", arg[2])
								m.Push("", value, []string{kit.MDB_NAME, kit.MDB_TEXT})
							})
						})
						break
					}
					m.Option("pod", "")
					m.Cmdy("web.label", arg[0], arg[1], "web.chat.search", "get", arg[2:])
					m.Sort("time", "time_r")
				}
			}},
			"commend": {Name: "commend label pod engine work auto", Help: "推荐引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					m.Cmdy("web.label", arg)
					return
				}

				switch arg[0] {
				case "add":
					if m.Richs(cmd, nil, arg[1], nil) == nil {
						m.Rich(cmd, nil, kit.Data(kit.MDB_NAME, arg[1]))
					}
					m.Richs(cmd, nil, arg[1], func(key string, value map[string]interface{}) {
						m.Grow(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(
							kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3:],
						))
					})
				case "get":
					wg := &sync.WaitGroup{}
					m.Richs(cmd, nil, arg[1], func(key string, value map[string]interface{}) {
						wg.Add(1)
						m.Gos(m, func(m *ice.Message) {
							m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
								m.Cmdy(value[kit.MDB_TEXT], arg[2:])
							})
							wg.Done()
						})
					})
					m.Sort("time", "time_r")
					wg.Wait()
				case "set":
					if arg[1] != "" {
						m.Cmdy("web.space", arg[1], "web.chat.commend", "set", "", arg[2:])
						break
					}

					if m.Richs(cmd, "meta.user", m.Option(ice.MSG_USERNAME), nil) == nil {
						m.Rich(cmd, "meta.user", kit.Dict(
							kit.MDB_NAME, m.Option(ice.MSG_USERNAME),
						))
					}
					switch m.Option("_action") {
					case "喜欢":
						m.Richs(cmd, "meta.user", m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
							m.Grow(cmd, kit.Keys("meta.user", kit.MDB_HASH, key, "like"), kit.Dict(
								kit.MDB_EXTRA, kit.Dict("engine", arg[2], "favor", arg[3], "id", arg[4]),
								kit.MDB_TYPE, arg[5], kit.MDB_NAME, arg[6], kit.MDB_TEXT, arg[7],
							))
						})
					case "讨厌":
						m.Richs(cmd, "meta.user", m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
							m.Grow(cmd, kit.Keys("meta.user", kit.MDB_HASH, key, "hate"), kit.Dict(
								kit.MDB_EXTRA, kit.Dict("engine", arg[2], "favor", arg[3], "id", arg[4]),
								kit.MDB_TYPE, arg[5], kit.MDB_NAME, arg[6], kit.MDB_TEXT, arg[7],
							))
						})
					}

					m.Richs(cmd, nil, arg[2], func(key string, value map[string]interface{}) {
						m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							m.Cmdy(value[kit.MDB_TEXT], "set", arg[3:])
						})
					})
				default:
					if len(arg) == 2 {
						m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
							m.Push(key, value)
						})
						break
					}
					if len(arg) == 3 {
						m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
							m.Push(key, value)
						})
						break
					}
					m.Cmdy("web.label", arg[0], arg[1], "web.chat.commend", "get", arg[2:])
					// m.Cmdy("web.chat.commend", "get", arg[2:])
				}
			}},
		}}, nil)
}
