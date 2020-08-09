package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

const (
	CHECK = "check"
	LOGIN = "login"
	TITLE = "title"
)
const HEADER = "header"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			HEADER: {Name: "header", Help: "标题栏", Value: kit.Dict(
				TITLE, "github.com/shylinux/contexts",
			)},
		},
		Commands: map[string]*ice.Command{
			"/" + HEADER: {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
				CHECK: {Name: "check", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},
				LOGIN: {Name: "login", Help: "用户登录", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},

				"pack": {Name: "pack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					if f, p, e := kit.Create("usr/volcanos/cache_data.js"); m.Assert(e) {
						defer f.Close()
						data := kit.UnMarshal(m.Option("content"))
						f.WriteString(`Volcanos.meta.pack = ` + kit.Formats(data))
						m.Echo(p)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(m.Conf(HEADER, TITLE))
			}},
		},
	}, nil)
}
