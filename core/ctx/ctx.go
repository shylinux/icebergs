package ctx

import (
	"github.com/shylinux/icebergs"
)

var Index = &ice.Context{Name: "ctx", Help: "元始模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		"hi": {Name: "hi", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
