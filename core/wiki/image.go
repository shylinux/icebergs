package wiki

import (
	ice "github.com/shylinux/icebergs"
)

func init() {

	Index.Register(&ice.Context{Name: "jpg", Help: "图片",
		Configs: map[string]*ice.Config{},
		Commands: map[string]*ice.Command{
			"list": {Name: "list name", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(arg[0])
			}},
			"save": {Name: "save name text", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			}},
			"show": {Name: "show name", Help: "渲染", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			}},
		},
	}, nil)

}
