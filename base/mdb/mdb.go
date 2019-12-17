package mdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
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
	},
}

func init() { ice.Index.Register(Index, nil) }
