package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/footer": {Name: "/footer", Help: "状态栏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			kit.Fetch(m.Confv(ice.WEB_SERVE, "meta.legal"), func(index int, value string) {
				m.Echo(value)
			})
		}},
	}}, nil)
}
