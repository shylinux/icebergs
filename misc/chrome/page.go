package crx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	kit "github.com/shylinux/toolkits"
)

const Page = "page"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			Page: {Name: "page", Help: "网页", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			"/page": {Name: "/page", Help: "网页", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == "get" {
						switch m.Option("hostname") {
						case "music.163.com":
							m.Option("top", "200")
							m.Result("web.code.chrome.spide", "", m.Option("tid"))
							return
						case "localhost", "fib.woa.com":
							return
						}
						return
						m.Option("top", "200")
						m.Echo("cli.runtime")
						return
					}
					if arg[0] == cli.RUN {
						m.Cmdy(arg[1:])
						return
					}
					m.Cmdy(ctx.COMMAND, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
		},
	})
}
