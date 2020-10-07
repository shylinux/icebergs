package yac

import (
	"github.com/shylinux/icebergs"
)

const YAC = "yac"

var Index = &ice.Context{Name: YAC, Help: "语法模块",
	Commands: map[string]*ice.Command{
		"hi": {Name: "hi", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
