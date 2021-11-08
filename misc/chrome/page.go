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
				m.Cmdy(STYLE, ctx.ACTION, ctx.COMMAND, arg)
				m.Cmdy(FIELD, ctx.ACTION, ctx.COMMAND, arg)
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FIELD, ice.RUN, arg)
			}},
		}},
	}})
}
