package team

import (
	ice "github.com/shylinux/icebergs"
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
				m.Echo("hello world")
			}},
			"项目测试": {Name: "项目测试", Help: "项目测试", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				m.Echo("hello world")
			}},
		},
	}, &web.Frame{})
}
