package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const POD = "pod"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/pod/": {Name: "/pod/", Help: "节点", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) == 0 {
						m.Push("index", CMD)
						m.Push("args", "")
						return
					}
					m.Cmdy(ctx.COMMAND, arg[0])
				}},
				cli.RUN: {Name: "command", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if kit.Select("", arg, 0) == "" {
					_cmd_render(m, web.ROUTE)
					return
				}
				m.RenderIndex(web.SERVE, ice.VOLCANOS)
			}},
		},
		Configs: map[string]*ice.Config{
			POD: {Name: POD, Help: "节点", Value: kit.Data()},
		},
	})
}
