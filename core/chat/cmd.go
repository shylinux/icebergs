package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CMD = "cmd"

func init() {
	Index.MergeCommands(ice.Commands{
		CMD: {Actions: web.ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			if len(arg[0]) == 0 || arg[0] == "" {
				web.RenderMain(m)
			} else if m.IsCliUA() {
				kit.If(aaa.Right(m, arg), func() { m.Cmdy(arg, m.Optionv(ice.ARG)).RenderResult() })
			} else if arg[0] == web.CHAT_PORTAL {
				web.RenderMain(m)
			} else if m.Cmdy(ctx.COMMAND, arg[0]); m.Length() > 0 {
				web.RenderCmd(m, m.Append(ctx.INDEX), arg[1:])
			}
		}},
	})
}
