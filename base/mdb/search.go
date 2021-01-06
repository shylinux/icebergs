package mdb

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const SEARCH = "search"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SEARCH: {Name: "search", Help: "搜索", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TYPE)},
		},
		Commands: map[string]*ice.Command{
			SEARCH: {Name: "search type word text auto", Help: "搜索", Action: map[string]*ice.Action{
				CREATE: {Name: "create type cmd ctx", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Rich(SEARCH, nil, kit.Dict(kit.MDB_TYPE, arg[0], kit.MDB_NAME, kit.Select(arg[0], arg, 1), kit.MDB_TEXT, kit.Select("", arg, 2)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || arg[0] == "" {
					m.Richs(SEARCH, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						m.Push(key, value, []string{kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
					})
					return
				}

				if strings.Contains(arg[1], ";") {
					arg = strings.Split(arg[1], ";")
				}
				m.Option(FIELDS, kit.Select("pod,ctx,cmd,time,size,type,name,text", kit.Select(m.Option(FIELDS), arg, 2)))
				for _, k := range strings.Split(arg[0], ",") {
					m.Richs(SEARCH, nil, k, func(key string, value map[string]interface{}) {
						m.Cmdy(kit.Keys(value[kit.MDB_TEXT], value[kit.MDB_NAME]), SEARCH, arg[0], arg[1], kit.Select("", arg, 2))
					})
				}
			}},
		}})
}
