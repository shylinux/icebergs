package chat

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const (
	LEGAL = "legal"
)
const FOOTER = "footer"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FOOTER: {Name: "footer", Help: "状态栏", Value: kit.Dict(
				LEGAL, []interface{}{`<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`},
			)},
		},
		Commands: map[string]*ice.Command{
			"/footer": {Name: "/footer", Help: "状态栏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				kit.Fetch(m.Confv(FOOTER, LEGAL), func(index int, value string) { m.Echo(value) })
			}},
		},
	})
}
