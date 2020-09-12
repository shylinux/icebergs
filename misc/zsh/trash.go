package zsh

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"net/url"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/help": {Name: "/help", Help: "帮助", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy("help")
			}},
			"/login": {Name: "/login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy("login", "init", c.Name)
			}},
			"/logout": {Name: "/logout", Help: "登出", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy("login", "exit")
			}},

			"/ish": {Name: "/ish", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if sub, e := url.QueryUnescape(m.Option("sub")); m.Assert(e) {
					m.Cmdy(kit.Split(sub))
					if len(m.Resultv()) == 0 {
						m.Table()
					}
				}
			}},
		},
	}, nil)
}
