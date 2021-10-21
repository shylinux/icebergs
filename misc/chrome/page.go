package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	kit "shylinux.com/x/toolkits"
)

const Page = "page"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		Page: {Name: "page", Help: "网页", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		"/page": {Name: "/page", Help: "网页", Action: map[string]*ice.Action{
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 {
					m.Cmdy(STYLE, SYNC, m.OptionSimple("hostname"), ice.OptionFields(""))
					m.Cmdy(FIELD, SYNC, m.OptionSimple("hostname"), ice.OptionFields(""))
					return
				}
				m.Cmdy(ctx.COMMAND, arg)
			}},
			FIELD: {Name: "field", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FIELD, arg)
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

		}},
	}})
}
