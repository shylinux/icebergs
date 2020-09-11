package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{},
		Commands: map[string]*ice.Command{
			"/favor": {Name: "/favor", Help: "收藏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 && arg[0] != "sh" {
					// 添加收藏
					m.Cmdy(web.FAVOR, kit.Select(m.Conf("zsh", "meta.history"), m.Option("tab")),
						kit.Select(web.TYPE_SHELL, m.Option("type")), m.Option("note"), arg[0])
					return
				}

				if m.Option("tab") == "" {
					// 收藏列表
					m.Cmdy(web.FAVOR).Table()
					return
				}

				m.Echo("#/bin/sh\n\n")
				m.Cmd(web.PROXY, m.Option("you"), web.FAVOR, m.Option("tab")).Table(func(index int, value map[string]string, head []string) {
					switch value["type"] {
					case web.TYPE_SHELL:
						// 查看收藏
						if m.Option("note") == "" || m.Option("note") == value["name"] {
							m.Echo("# %v\n%v\n\n", value["name"], value["text"])
						}
					}
				})
			}},
		},
	}, nil)
}
