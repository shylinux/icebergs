package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
	"sync"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			"search": {Name: "search", Help: "search", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			"search": {Name: "search label pod engine word", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					m.Cmdy(web.LABEL, arg)
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
						m.Cmdy(web.SPACE, arg[1], "web.chat.search", "set", "", arg[2:])
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
					m.Cmdy(web.LABEL, arg[0], arg[1], "web.chat.search", "get", arg[2:])
					m.Sort("time", "time_r")
				}
			}},
		}}, nil)
}
