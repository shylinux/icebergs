package lex

import (
	ice "github.com/shylinux/icebergs"
)

const LEX = "lex"

var Index = &ice.Context{Name: LEX, Help: "词法模块",
	Commands: map[string]*ice.Command{
		"hi": {Name: "hi", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
