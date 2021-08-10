package chrome

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	kit "github.com/shylinux/toolkits"
)

const Page = "page"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		Page: {Name: "page", Help: "网页", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		"/page": {Name: "/page", Help: "网页", Action: map[string]*ice.Action{
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 {
					m.Cmdy(STYLE, SYNC, m.OptionSimple("hostname"))

					switch m.Option("hostname") {
					case "golang.google.cn":
						m.Option("top", "200")
						m.Option("selection", "word")
						m.Result("web.wiki.alpha.alpha")

					case "music.163.com":
						m.Option("top", "200")
						m.Result(SPIDE, "", m.Option("tid"))
					case "localhost", "fib.woa.com":
						return
					}
					return
					m.Option("top", "200")
					m.Echo("cli.runtime")
					return
				}
				m.Cmdy(ctx.COMMAND, arg)
			}},
			cli.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

		}},
	}})
}
