package team

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
)

func _work_input(m *ice.Message, key, value string) {
	switch key {
	case "name":
		m.Push(key, "volcanos")
		m.Push(key, "icebergs")
		m.Push(key, "intshell")
		m.Push(key, "contexts")
		m.Push(key, "toolkits")
		m.Push(key, "learning")
	case "text":
		m.Push(key, "volcanos")
		m.Push(key, "icebergs")
	}
}
func init() {
	Index.Register(&ice.Context{Name: "工作", Help: "工作",
		Commands: map[string]*ice.Command{
			"项目开发": {Name: "项目开发", Help: "项目开发", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {

			}},
			"项目测试": {Name: "项目测试", Help: "项目测试", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {

			}},
		},
	}, &web.Frame{})

	Index.Register(&ice.Context{Name: "创业", Help: "创业",
		Commands: map[string]*ice.Command{
			"项目调研": {Name: "项目调研", Help: "项目调研",
				Action: map[string]*ice.Action{
					"input": {Name: "input", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
						_work_input(m, arg[0], arg[1])
					}},
				},
				Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {

				}},
			"项目开发": {Name: "项目开发", Help: "项目开发",
				Action: map[string]*ice.Action{
					"input": {Name: "input", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
						_work_input(m, arg[0], arg[1])
					}},
				},
				Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
					m.Echo("hello world")
				}},
			"项目测试": {Name: "项目测试", Help: "项目测试", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {

			}},
		},
	}, &web.Frame{})
}
